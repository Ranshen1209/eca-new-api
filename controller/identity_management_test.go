package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/service/authz"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupIdentityManagementTest(t *testing.T) (*model.User, service.AuthIdentity) {
	t.Helper()
	user, identity := setupSecurityEnrollmentTest(t)
	logConnection, err := model.LOG_DB.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, logConnection.Close()) })
	require.NoError(t, model.DB.AutoMigrate(&model.CasbinRule{}, &model.AuthzRole{}))
	require.NoError(t, authz.Init(model.DB))
	return user, identity
}

func TestOIDCIdentityEndpointsRejectLocalManagement(t *testing.T) {
	user, identity := setupIdentityManagementTest(t)
	require.NoError(t, model.DB.Model(user).Update("oidc_id", "synthetic-oidc-subject").Error)
	endpoints := []struct {
		name, method, body string
		handler            gin.HandlerFunc
	}{
		{"username", "PUT", `{"username":"changed"}`, UpdateSelf},
		{"email_with_preference", "PUT", `{"email":"changed@example.test","language":"zh"}`, UpdateSelf},
		{"password", "PUT", `{"username":"enrollment-user","password":"changed-password"}`, UpdateSelf},
		{"password_with_sidebar", "PUT", `{"sidebar_modules":"{}","password":"changed-password"}`, UpdateSelf},
		{"email_start", "POST", `{}`, EmailBindStart}, {"email_resend", "POST", `{}`, EmailBindResend}, {"email_finish", "POST", `{}`, EmailBind},
		{"totp_status", "GET", ``, Get2FAStatus}, {"totp_setup", "POST", `{}`, Setup2FA}, {"totp_finish", "POST", `{}`, Enable2FA}, {"totp_disable", "POST", `{}`, Disable2FA}, {"totp_backup", "POST", `{}`, RegenerateBackupCodes},
		{"passkey_status", "GET", ``, PasskeyStatus}, {"passkey_begin", "POST", `{}`, PasskeyRegisterBegin}, {"passkey_finish", "POST", `{}`, PasskeyRegisterFinish}, {"passkey_delete", "DELETE", ``, PasskeyDelete}, {"passkey_verify_begin", "POST", `{}`, PasskeyVerifyBegin}, {"passkey_verify_finish", "POST", `{}`, PasskeyVerifyFinish},
		{"wechat", "POST", `{}`, WeChatBind}, {"unbind_custom", "DELETE", ``, UnbindCustomOAuth},
		{"oauth_bind_start", "POST", `{"provider":"github","intent":"bind"}`, GenerateOAuthCode},
		{"delete_self", "DELETE", ``, DeleteSelf},
	}
	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			response := securityEnrollmentRequest(endpoint.method, "/api/test", endpoint.body, "", identity, endpoint.handler)
			require.Equal(t, http.StatusForbidden, response.Code, response.Body.String())
			require.Contains(t, response.Body.String(), "IDENTITY_MANAGED")
		})
	}
	admins := []struct {
		name, body string
		handler    gin.HandlerFunc
	}{
		{"username", fmt.Sprintf(`{"id":%d,"username":"changed"}`, user.Id), UpdateUser},
		{"password", fmt.Sprintf(`{"id":%d,"username":"enrollment-user","password":"changed-password"}`, user.Id), UpdateUser},
		{"email", fmt.Sprintf(`{"id":%d,"username":"enrollment-user","email":"changed@example.test"}`, user.Id), UpdateUser},
		{"binding", ``, AdminClearUserBinding}, {"custom_binding", ``, UnbindCustomOAuthByAdmin}, {"totp", ``, AdminDisable2FA}, {"passkey", ``, AdminResetPasskey},
	}
	for _, endpoint := range admins {
		t.Run("admin_"+endpoint.name, func(t *testing.T) {
			response := securityEnrollmentRequest("PUT", "/api/user", endpoint.body, "", identity, func(c *gin.Context) {
				c.Set("role", common.RoleRootUser)
				c.Params = gin.Params{{Key: "id", Value: fmt.Sprint(user.Id)}, {Key: "provider_id", Value: "1"}, {Key: "binding_type", Value: "oidc"}}
				endpoint.handler(c)
			})
			require.Equal(t, http.StatusForbidden, response.Code, response.Body.String())
		})
	}
	unchanged, err := model.GetUserById(user.Id, true)
	require.NoError(t, err)
	require.Equal(t, user.Username, unchanged.Username)
	require.Equal(t, user.Password, unchanged.Password)
	require.Equal(t, "synthetic-oidc-subject", unchanged.OidcId)
}

func TestOIDCIdentityRecoveryAndLogin(t *testing.T) {
	user, identity := setupIdentityManagementTest(t)
	require.NoError(t, model.DB.Model(user).Updates(map[string]any{"oidc_id": "synthetic-oidc-subject", "email": "employee@example.test"}).Error)
	common.RegisterVerificationCodeWithKey("employee@example.test", "synthetic-reset", common.PasswordResetPurpose)
	t.Cleanup(func() { common.DeleteKey("employee@example.test", common.PasswordResetPurpose) })
	// A reset link issued before OIDC enrollment cannot change the password.
	response := securityEnrollmentRequest("POST", "/api/user/reset", `{"email":"employee@example.test","token":"synthetic-reset"}`, "", identity, ResetPassword)
	var result map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Equal(t, false, result["success"])
	require.ErrorIs(t, model.ResetUserPasswordByEmail("employee@example.test", "new-password"), model.ErrIdentityManaged)
	current, err := model.GetUserById(user.Id, true)
	require.NoError(t, err)
	require.Equal(t, user.Password, current.Password)
	challenge, err := service.StartLoginVerification(current, "oauth:oidc", nil)
	require.NoError(t, err)
	require.Nil(t, challenge)
	challenge, err = service.StartLoginVerification(current, "oidc", nil)
	require.NoError(t, err)
	require.Nil(t, challenge)
	for _, method := range []string{"password", "passkey", "oauth:github"} {
		_, err := service.StartLoginVerification(current, method, nil)
		require.ErrorIs(t, err, model.ErrIdentityManaged)
	}
	response = securityEnrollmentRequest("GET", "/api/reset_password?email=employee@example.test", "", "", identity, SendPasswordResetEmail)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), `"success":true`)
	require.True(t, common.VerifyCodeWithKey("employee@example.test", "synthetic-reset", common.PasswordResetPurpose))
}

func TestOIDCIdentityPreservesPreferencesAndOperationalAccounts(t *testing.T) {
	user, identity := setupIdentityManagementTest(t)
	// Unbound operational accounts retain password verification.
	requirements, err := service.GetVerificationRequirements(identity, service.VerificationScopePasswordChange)
	require.NoError(t, err)
	require.Equal(t, service.VerificationMethodPassword, requirements.Methods[0].Method)
	require.NoError(t, model.DB.Model(user).Update("oidc_id", "synthetic-oidc-subject").Error)
	response := securityEnrollmentRequest("PUT", "/api/user/self", `{"language":"zh"}`, "", identity, UpdateSelf)
	require.Contains(t, response.Body.String(), `"success":true`)
	current, err := model.GetUserById(user.Id, false)
	require.NoError(t, err)
	require.Equal(t, "zh", current.GetSetting().Language)
	require.Equal(t, true, buildSelfUserData(current)["identity_managed"])
	// App permissions/group/display name remain editable; no identity mutation.
	response = securityEnrollmentRequest("PUT", "/api/user", fmt.Sprintf(`{"id":%d,"username":"enrollment-user","display_name":"App nickname","group":"default","remark":"business note"}`, user.Id), "", identity, func(c *gin.Context) { c.Set("role", common.RoleRootUser); UpdateUser(c) })
	require.Contains(t, response.Body.String(), `"success":true`)
	current, err = model.GetUserById(user.Id, false)
	require.NoError(t, err)
	require.Equal(t, "App nickname", current.DisplayName)
	for _, scope := range []string{service.VerificationScopePasswordChange, service.VerificationScopeTwoFASetup, service.VerificationScopePasskeyRegister, service.VerificationScopeAccountBind} {
		_, err := service.GetVerificationRequirements(identity, scope)
		require.ErrorIs(t, err, model.ErrIdentityManaged)
	}
}

func TestOIDCIdentityPreservesBusinessAccessTokens(t *testing.T) {
	user, identity := setupIdentityManagementTest(t)
	require.NoError(t, model.DB.Model(user).Update("oidc_id", "synthetic-oidc-subject").Error)
	settings := system_setting.GetOIDCSettings()
	original := *settings
	t.Cleanup(func() { *settings = original })
	settings.Enabled = true
	requirements, err := service.GetVerificationRequirements(identity, service.VerificationScopeAccessTokenGenerate)
	require.NoError(t, err)
	require.Equal(t, []service.VerificationMethodOption{{Method: service.VerificationMethodOAuth, Available: true}}, requirements.Methods)
	require.Len(t, requirements.OAuthProviders, 1)
	require.Equal(t, "oidc", requirements.OAuthProviders[0].Slug)
	proof := issueSecurityEnrollmentProof(t, identity, service.VerificationOperation{Scope: service.VerificationScopeAccessTokenGenerate, Context: []byte(`{"scopes":["profile:read"],"expires_at":0}`)}, service.VerificationMethodOAuth)
	response := securityEnrollmentRequest("POST", "/api/user/access_tokens", `{"name":"business-automation","scopes":["profile:read"],"expires_at":0}`, proof, identity, CreateAccessToken)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), `"success":true`)
	tokens, err := model.ListUserAccessTokens(user.Id)
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	require.Equal(t, "business-automation", tokens[0].Name)
	require.Equal(t, []string{"profile:read"}, tokens[0].GetScopes())
}
