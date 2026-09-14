package service

import (
	"context"
	"math"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEcaAllowance(t *testing.T) {
	t.Helper()
	truncate(t)
	require.NoError(t, model.DB.AutoMigrate(&model.EcaQuotaGrant{}))
	require.NoError(t, model.DB.Exec("DELETE FROM eca_quota_grants").Error)
	t.Cleanup(func() { require.NoError(t, model.DB.Exec("DELETE FROM eca_quota_grants").Error) })
	settings := operation_setting.GetGeneralSetting()
	oldDisplay, oldRate, oldUnit := settings.QuotaDisplayType, operation_setting.USDExchangeRate, common.QuotaPerUnit
	t.Cleanup(func() {
		settings.QuotaDisplayType, operation_setting.USDExchangeRate, common.QuotaPerUnit = oldDisplay, oldRate, oldUnit
	})
	settings.QuotaDisplayType = "CNY"
	operation_setting.USDExchangeRate = 6.72
	common.QuotaPerUnit = 500000
	t.Setenv("ECA_AUTO_QUOTA_ENABLED", "true")
	t.Setenv("ECA_AUTO_QUOTA_THRESHOLD_CNY", "")
	t.Setenv("ECA_AUTO_QUOTA_CREDIT_CNY", "")
	t.Setenv("BATCH_UPDATE_ENABLED", "false")
}

func TestEcaAutoQuotaCNYConversionAndValidation(t *testing.T) {
	setupEcaAllowance(t)
	threshold, credit, err := ecaAutoQuotaAmounts()
	require.NoError(t, err)
	assert.Equal(t, 7440476, threshold)
	assert.Equal(t, 37202381, credit)
	for _, tc := range []struct{ threshold, credit string }{{"-1", "500"}, {"NaN", "500"}, {"100", "0"}, {"100", "100"}, {"100", "99999999999999999999999"}} {
		t.Run(tc.threshold+"/"+tc.credit, func(t *testing.T) {
			t.Setenv("ECA_AUTO_QUOTA_THRESHOLD_CNY", tc.threshold)
			t.Setenv("ECA_AUTO_QUOTA_CREDIT_CNY", tc.credit)
			_, _, err := ecaAutoQuotaAmounts()
			require.Error(t, err)
		})
	}
	for _, rate := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		operation_setting.USDExchangeRate = rate
		_, _, err := ecaAutoQuotaAmounts()
		require.Error(t, err)
	}
	operation_setting.USDExchangeRate = 6.72
	operation_setting.GetGeneralSetting().QuotaDisplayType = "USD"
	_, _, err = ecaAutoQuotaAmounts()
	require.Error(t, err)
}

func TestEcaAutoQuotaTaskAndFutureUsers(t *testing.T) {
	setupEcaAllowance(t)
	user := model.User{Username: "eca-allowance", Password: "unused", Status: 1, Role: 1, Quota: 7440476, UsedQuota: 123, RequestCount: 7}
	require.NoError(t, model.DB.Create(&user).Error)
	task, err := model.CreateSystemTask("eca_auto_quota", nil, nil)
	require.NoError(t, err)
	_, claimed, err := model.ClaimSystemTask(task.ID, task.Type, "eca-test", common.GetTimestamp()+60)
	require.NoError(t, err)
	require.True(t, claimed)
	(ecaAutoQuotaHandler{}).Run(context.Background(), task, "eca-test")
	finished, err := model.GetSystemTaskByTaskID(task.TaskID)
	require.NoError(t, err)
	assert.Equal(t, model.SystemTaskStatusSucceeded, finished.Status)
	assert.JSONEq(t, `{"credited_users":1}`, finished.Result)
	var after model.User
	require.NoError(t, model.DB.First(&after, user.Id).Error)
	assert.Equal(t, 44642857, after.Quota)
	assert.Equal(t, 123, after.UsedQuota)
	assert.Equal(t, 7, after.RequestCount)
	var logs int64
	require.NoError(t, model.LOG_DB.Model(&model.Log{}).Where("user_id = ? AND type = ?", user.Id, model.LogTypeSystem).Count(&logs).Error)
	assert.Equal(t, int64(1), logs)
	count, err := runEcaAutoQuota(context.Background(), task.CreatedAt/60)
	require.NoError(t, err)
	assert.Zero(t, count)
	newUser := model.User{Username: "eca-new-employee", Password: "unused", Status: 1, Role: 1, Quota: 0, AffCode: "eca-new"}
	require.NoError(t, model.DB.Create(&newUser).Error)
	count, err = runEcaAutoQuota(context.Background(), task.CreatedAt/60+1)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestEcaAutoQuotaDisabledAndBatchMode(t *testing.T) {
	setupEcaAllowance(t)
	user := model.User{Username: "eca-disabled-task", Password: "unused", Status: 1, Role: 1, Quota: 50}
	require.NoError(t, model.DB.Create(&user).Error)
	t.Setenv("ECA_AUTO_QUOTA_ENABLED", "false")
	count, err := runEcaAutoQuota(context.Background(), 1000)
	require.NoError(t, err)
	assert.Zero(t, count)
	assert.False(t, (ecaAutoQuotaHandler{}).Enabled())
	t.Setenv("ECA_AUTO_QUOTA_ENABLED", "true")
	t.Setenv("BATCH_UPDATE_ENABLED", "true")
	_, err = runEcaAutoQuota(context.Background(), 1000)
	require.ErrorContains(t, err, "BATCH_UPDATE_ENABLED")
	var after model.User
	require.NoError(t, model.DB.First(&after, user.Id).Error)
	assert.Equal(t, 50, after.Quota)
}
