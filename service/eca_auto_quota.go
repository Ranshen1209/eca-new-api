package service

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
)

type ecaAutoQuotaHandler struct{}

func init() {
	RegisterSystemTaskHandler(ecaAutoQuotaHandler{})
}

func (ecaAutoQuotaHandler) Type() string { return "eca_auto_quota" }
func (ecaAutoQuotaHandler) Enabled() bool {
	return os.Getenv("ECA_AUTO_QUOTA_ENABLED") == "true"
}
func (ecaAutoQuotaHandler) Interval() time.Duration { return time.Minute }
func (ecaAutoQuotaHandler) NewPayload() any         { return nil }

// Resolve CNY amounts using the same rate and quota units as the site. Reject
// incompatible configuration instead of crediting a different currency silently.
func ecaAutoQuotaAmounts() (int, int, error) {
	rate := operation_setting.USDExchangeRate
	if !operation_setting.IsCNYDisplay() || rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) ||
		common.QuotaPerUnit <= 0 || math.IsNaN(common.QuotaPerUnit) || math.IsInf(common.QuotaPerUnit, 0) {
		return 0, 0, fmt.Errorf("internal allowance requires CNY display and a positive finite exchange rate and quota unit")
	}
	thresholdText := os.Getenv("ECA_AUTO_QUOTA_THRESHOLD_CNY")
	if thresholdText == "" {
		thresholdText = "100"
	}
	creditText := os.Getenv("ECA_AUTO_QUOTA_CREDIT_CNY")
	if creditText == "" {
		creditText = "500"
	}
	threshold, err := decimal.NewFromString(thresholdText)
	if err != nil || threshold.IsNegative() {
		return 0, 0, fmt.Errorf("invalid ECA_AUTO_QUOTA_THRESHOLD_CNY")
	}
	credit, err := decimal.NewFromString(creditText)
	if err != nil || !credit.GreaterThan(threshold) {
		return 0, 0, fmt.Errorf("ECA_AUTO_QUOTA_CREDIT_CNY must exceed the nonnegative threshold")
	}
	units := decimal.NewFromFloat(common.QuotaPerUnit).Div(decimal.NewFromFloat(rate))
	threshold = threshold.Mul(units).Floor()
	credit = credit.Mul(units).Round(0)
	if credit.LessThan(decimal.NewFromInt(1)) || threshold.Add(credit).GreaterThan(decimal.NewFromInt(common.MaxWalletQuota)) {
		return 0, 0, fmt.Errorf("internal allowance exceeds wallet quota bounds")
	}
	return int(threshold.IntPart()), int(credit.IntPart()), nil
}

func (handler ecaAutoQuotaHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	count, err := runEcaAutoQuota(ctx, task.CreatedAt/60)
	status := model.SystemTaskStatusSucceeded
	errorMessage := ""
	if err != nil {
		status = model.SystemTaskStatusFailed
		errorMessage = err.Error()
		logger.LogWarn(ctx, "internal allowance task failed: "+errorMessage)
	}
	if err := model.FinishSystemTask(task.TaskID, runnerID, status, map[string]int{"credited_users": count}, errorMessage); err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("internal allowance task result could not be saved: %v", err))
	}
}

func runEcaAutoQuota(ctx context.Context, cycle int64) (int, error) {
	if !(ecaAutoQuotaHandler{}).Enabled() {
		return 0, nil
	}
	if common.BatchUpdateEnabled || os.Getenv("BATCH_UPDATE_ENABLED") == "true" {
		return 0, fmt.Errorf("internal allowance requires BATCH_UPDATE_ENABLED to be disabled")
	}
	threshold, credit, err := ecaAutoQuotaAmounts()
	if err != nil {
		return 0, err
	}
	count, lastID := 0, 0
	for {
		var ids []int
		if err := model.DB.WithContext(ctx).Model(&model.User{}).
			Where("id > ? AND status = ? AND quota <= ?", lastID, common.UserStatusEnabled, threshold).
			Order("id").Limit(200).Pluck("id", &ids).Error; err != nil {
			return count, err
		}
		if len(ids) == 0 {
			return count, nil
		}
		for _, id := range ids {
			grant, err := model.GrantEcaUserQuota(ctx, id, cycle, threshold, credit)
			if err != nil {
				return count, err
			}
			if grant != nil {
				count++
				// The grant table remains authoritative even if a separate log
				// database is temporarily unavailable after the credit commits.
				model.RecordLog(id, model.LogTypeSystem, fmt.Sprintf("公司内部自动补额（非实际支付）：增加 %s，余额 %s → %s", logger.LogQuota(credit), logger.LogQuota(grant.QuotaBefore), logger.LogQuota(grant.QuotaAfter)))
				logger.LogInfo(ctx, fmt.Sprintf("internal allowance granted: user_id=%d cycle=%d credit=%d before=%d after=%d", id, cycle, credit, grant.QuotaBefore, grant.QuotaAfter))
			}
			lastID = id
		}
	}
}
