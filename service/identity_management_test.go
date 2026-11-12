package service

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/require"
)

func TestOIDCIdentityVerificationUsesOnlyIdP(t *testing.T) {
	state := model.UserVerificationState{IdentityManaged: true, HasPassword: true, HasTwoFA: true, HasPasskey: true}
	for _, scope := range []string{VerificationScopeAccessTokenGenerate, VerificationScopeAccessTokenUpdate, VerificationScopeAccessTokenRevoke, VerificationScopeChannelKeyRead, VerificationScopeAdminUserUpdate} {
		methods, err := securityVerificationPolicy(scope, state)
		require.NoError(t, err)
		require.Equal(t, []VerificationMethodOption{{Method: VerificationMethodOAuth, Available: true}}, methods)
	}
	_, err := securityVerificationPolicy("unknown", state)
	require.ErrorIs(t, err, ErrProofScope)
	require.ErrorIs(t, requireLoginVerificationMethod(&state, VerificationMethodTwoFA), model.ErrIdentityManaged)
}
