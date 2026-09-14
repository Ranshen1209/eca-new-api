package model

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupEcaQuotaTest(t *testing.T) {
	t.Helper()
	truncateTables(t)
	require.NoError(t, DB.AutoMigrate(&EcaQuotaGrant{}))
	require.NoError(t, DB.Exec("DELETE FROM eca_quota_grants").Error)
	t.Cleanup(func() { require.NoError(t, DB.Exec("DELETE FROM eca_quota_grants").Error) })
}

func TestEcaQuotaEligibilityAndAccounting(t *testing.T) {
	for _, tc := range []struct {
		name                string
		quota, status, role int
		deleted, eligible   bool
	}{
		{"threshold", 100, 1, 1, false, true},
		{"below", 92, 1, 1, false, true},
		{"zero", 0, 1, 1, false, true},
		{"overdrawn", -800, 1, 1, false, true},
		{"above", 101, 1, 1, false, false},
		{"disabled", 20, common.UserStatusDisabled, 1, false, false},
		{"deleted", 20, 1, 1, true, false},
		{"admin", 20, 1, common.RoleAdminUser, false, true},
		{"root", 20, 1, common.RoleRootUser, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setupEcaQuotaTest(t)
			user := createReserveTestUser(t, tc.quota)
			require.NoError(t, DB.Model(&user).Updates(map[string]any{"used_quota": 123, "request_count": 7, "status": tc.status, "role": tc.role}).Error)
			if tc.deleted {
				require.NoError(t, DB.Delete(&user).Error)
			}
			grant, err := GrantEcaUserQuota(context.Background(), user.Id, 1000, 100, 500)
			require.NoError(t, err)
			var after User
			require.NoError(t, DB.Unscoped().First(&after, user.Id).Error)
			assert.Equal(t, 123, after.UsedQuota)
			assert.Equal(t, 7, after.RequestCount)
			assert.Equal(t, tc.status, after.Status)
			assert.Equal(t, tc.role, after.Role)
			if tc.eligible {
				require.NotNil(t, grant)
				assert.Equal(t, tc.quota+500, after.Quota)
				assert.Equal(t, tc.quota, grant.QuotaBefore)
				assert.Equal(t, after.Quota, grant.QuotaAfter)
			} else {
				assert.Nil(t, grant)
				assert.Equal(t, tc.quota, after.Quota)
			}
			var count int64
			require.NoError(t, DB.Model(&EcaQuotaGrant{}).Count(&count).Error)
			assert.Equal(t, tc.eligible, count == 1)
		})
	}
}

func TestEcaQuotaConcurrentRetryPreservesDebits(t *testing.T) {
	setupEcaQuotaTest(t)
	user := createReserveTestUser(t, -10000)
	var credited atomic.Int32
	var wg sync.WaitGroup
	errs := make(chan error, 32)
	for range 16 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			grant, err := GrantEcaUserQuota(context.Background(), user.Id, 1000, 100, 500)
			if grant != nil {
				credited.Add(1)
			}
			errs <- err
		}()
		go func() {
			defer wg.Done()
			errs <- DB.Model(&User{}).Where("id = ?", user.Id).Update("quota", gorm.Expr("quota - ?", 7)).Error
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	assert.Equal(t, int32(1), credited.Load())
	assert.Equal(t, -10000+500-16*7, getUserQuotaFromDB(t, user.Id))
	grant, err := GrantEcaUserQuota(context.Background(), user.Id, 1001, 100, 500)
	require.NoError(t, err)
	require.NotNil(t, grant)
	assert.Equal(t, -10000+1000-16*7, getUserQuotaFromDB(t, user.Id))
}

func TestEcaQuotaAuditFailureRollsBackCredit(t *testing.T) {
	setupEcaQuotaTest(t)
	user := createReserveTestUser(t, 92)
	require.NoError(t, DB.Callback().Create().Before("gorm:create").Register("eca_test_audit_failure", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Table == "eca_quota_grants" {
			tx.AddError(errors.New("audit unavailable"))
		}
	}))
	t.Cleanup(func() { require.NoError(t, DB.Callback().Create().Remove("eca_test_audit_failure")) })
	grant, err := GrantEcaUserQuota(context.Background(), user.Id, 1000, 100, 500)
	require.ErrorContains(t, err, "audit unavailable")
	assert.Nil(t, grant)
	assert.Equal(t, 92, getUserQuotaFromDB(t, user.Id))
	var count int64
	require.NoError(t, DB.Model(&EcaQuotaGrant{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestEcaQuotaCreditSynchronizesExistingCacheOnlyOnce(t *testing.T) {
	setupEcaQuotaTest(t)
	useUserCacheMiniRedis(t)
	user := createReserveTestUser(t, -800)
	_, err := GetUserCache(user.Id)
	require.NoError(t, err)
	grant, err := GrantEcaUserQuota(context.Background(), user.Id, 1000, 100, 500)
	require.NoError(t, err)
	require.NotNil(t, grant)
	grant, err = GrantEcaUserQuota(context.Background(), user.Id, 1000, 100, 500)
	require.NoError(t, err)
	assert.Nil(t, grant)
	quota, err := GetUserQuota(user.Id, false)
	require.NoError(t, err)
	assert.Equal(t, -300, quota)
	assert.Equal(t, -300, getUserQuotaFromDB(t, user.Id))
}

func TestEcaQuotaRejectsUnsafeInputAndCancelledWork(t *testing.T) {
	setupEcaQuotaTest(t)
	user := createReserveTestUser(t, 92)
	for _, pair := range [][2]int{{-1, 500}, {100, 0}, {100, -500}, {100, common.MaxWalletQuota}, {common.MaxWalletQuota, 1}} {
		grant, err := GrantEcaUserQuota(context.Background(), user.Id, 1000, pair[0], pair[1])
		require.Error(t, err)
		assert.Nil(t, grant)
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	_, err := GrantEcaUserQuota(ctx, user.Id, 1000, 100, 500)
	require.Error(t, err)
	assert.Equal(t, 92, getUserQuotaFromDB(t, user.Id))
}
