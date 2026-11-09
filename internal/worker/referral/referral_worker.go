package referralworker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
)

type Worker struct {
	referralRepo interfaces.ReferralRepository
	uow          interfaces.UnitOfWork

	interval time.Duration
	log      *slog.Logger
}

func NewWorker(referralRepo interfaces.ReferralRepository, uow interfaces.UnitOfWork, interval time.Duration, log *slog.Logger) *Worker {
	if interval <= 0 {
		interval = 4 * time.Hour
	}
	return &Worker{
		referralRepo: referralRepo,
		uow:          uow,
		interval:     interval,
		log:          log,
	}
}
func (w *Worker) processPendingReferrals(ctx context.Context) {
	refferals, err := w.referralRepo.GetPendingReferrals(ctx, 100)
	if err != nil {
		w.log.Error("failed to fetch pending referrals", "error", err)
		return
	}
	for _, referral := range refferals {
		if ctx.Err() != nil {
			return
		}
		err := w.uow.WithinTransaction(ctx, func(repos interfaces.TransactionRepositories) error {
			ref, err := repos.Referral.GetByIDForUpdate(ctx, referral.ID)
			if err != nil {
				return err
			}
			if ref.Status != models.ReferralStatusPending {
				return nil
			}
			eligible, err := repos.Order.HasQualifyingDeliveredOrder(ctx, ref.ReferredUserID, ref.MinimumOrderAmount)
			if err != nil {
				return err
			}
			if !eligible {
				return nil
			}
			creditReward := func(userID uint, amount float64, rewardStatus *models.ReferralRewardStatus) error {
				if *rewardStatus == models.ReferralRewardStatusCompleted {
					return nil
				}
				if amount > 0 {
					wallet, err := repos.Wallet.GetByUserIDForUpdate(ctx, userID)
					if err != nil {
						return err
					}
					balanceBefore := wallet.Balance
					wallet.Balance += amount
					wallet.UpdatedAt = time.Now()
					if err := repos.Wallet.Update(ctx, wallet); err != nil {
						return err
					}
					description := fmt.Sprintf("Referral reward for referral #%d", ref.ID)
					referralID := ref.ID
					transaction := &models.WalletTransaction{
						WalletID:        wallet.ID,
						TransactionType: models.WalletTransactionCredit,
						Source:          models.WalletSourceReferralReward,
						Amount:          amount,
						BalanceBefore:   balanceBefore,
						BalanceAfter:    wallet.Balance,
						ReferralID:      &referralID,
						Description:     &description,
					}
					if err := repos.Wallet.CreateTransaction(ctx, transaction); err != nil {
						return err
					}
				}
				*rewardStatus = models.ReferralRewardStatusCompleted
				return nil
			}
			if err := creditReward(ref.ReferrerID, ref.ReferrerReward, &ref.ReferrerRewardStatus); err != nil {
				return err
			}
			if err := creditReward(ref.ReferredUserID, ref.ReferredReward, &ref.ReferredRewardStatus); err != nil {
				return err
			}
			if ref.ReferredRewardStatus == models.ReferralRewardStatusCompleted &&
				ref.ReferrerRewardStatus == models.ReferralRewardStatusCompleted {
				now := time.Now()
				ref.Status = models.ReferralStatusCompleted
				ref.CompletedAt = &now
				if err := repos.Referral.Update(ctx, ref); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			w.log.Error(
				"failed to process referral rewards",
				"referral_id", referral.ID,
				"error", err,
			)
			continue
		}
		w.log.Info(
			"referral processing completed",
			"referral_id", referral.ID,
		)
	}

}
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	w.log.Info("referral reward worker started", "interval", w.interval)
	w.processPendingReferrals(ctx)
	for {
		select {
		case <-ctx.Done():
			w.log.Info("referral reward worker stopped")
			return
		case <-ticker.C:
			w.processPendingReferrals(ctx)
		}
	}
}

var _ interfaces.Worker = (*Worker)(nil)
