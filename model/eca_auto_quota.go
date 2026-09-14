package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// EcaQuotaGrant is the durable record of an internal allowance, not a payment.
// One grant per user per UTC minute also fences retries after a process restart.
type EcaQuotaGrant struct {
	UserID      int   `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	Cycle       int64 `json:"cycle" gorm:"primaryKey;autoIncrement:false"`
	CreatedAt   int64 `json:"created_at" gorm:"bigint;index"`
	Threshold   int   `json:"threshold" gorm:"type:bigint"`
	Credit      int   `json:"credit" gorm:"type:bigint"`
	QuotaBefore int   `json:"quota_before" gorm:"type:bigint"`
	QuotaAfter  int   `json:"quota_after" gorm:"type:bigint"`
}

var errEcaQuotaGrantSkipped = errors.New("internal quota grant skipped")

// GrantEcaUserQuota checks eligibility and writes both the credit and its audit
// record in one transaction. It never replaces a balance read before the write.
func GrantEcaUserQuota(ctx context.Context, userID int, cycle int64, threshold, credit int) (*EcaQuotaGrant, error) {
	if userID <= 0 || cycle <= 0 || threshold < 0 || credit <= 0 ||
		credit > common.MaxWalletQuota || threshold > common.MaxWalletQuota-credit {
		return nil, fmt.Errorf("invalid internal quota grant parameters")
	}
	if common.BatchUpdateEnabled {
		return nil, fmt.Errorf("internal quota grants require synchronous wallet updates")
	}
	grant := EcaQuotaGrant{UserID: userID, Cycle: cycle, CreatedAt: common.GetTimestamp(), Threshold: threshold, Credit: credit}
	err := DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// This guarded write serializes competing grants and deductions on the
		// user row (SQLite serializes writers). Subsequent failures roll it back.
		result := tx.Model(&User{}).
			Where("id = ? AND status = ? AND quota <= ?", userID, common.UserStatusEnabled, threshold).
			Update("quota", gorm.Expr("quota + ?", credit))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errEcaQuotaGrantSkipped
		}
		var count int64
		if err := tx.Model(&EcaQuotaGrant{}).Where("user_id = ? AND cycle = ?", userID, cycle).Count(&count).Error; err != nil {
			return err
		}
		if count != 0 {
			return errEcaQuotaGrantSkipped
		}
		if err := tx.Model(&User{}).Where("id = ?", userID).Select("quota").Scan(&grant.QuotaAfter).Error; err != nil {
			return err
		}
		grant.QuotaBefore = grant.QuotaAfter - credit
		return tx.Create(&grant).Error
	})
	if errors.Is(err, errEcaQuotaGrantSkipped) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	syncCreditUserQuotaCache(userID, credit, "eca internal allowance")
	return &grant, nil
}
