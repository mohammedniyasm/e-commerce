package coupon

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
)

type Worker struct {
	couponRepo          interfaces.CouponRepository
	notificationRepo    interfaces.NotificationRepository
	couponUsageRepo     interfaces.CouponUsageRepository
	notificationUseCase interfaces.NotificationUseCase

	interval time.Duration
	log      *slog.Logger
}

func NewWorker(
	couponRepo interfaces.CouponRepository,
	notificationRepo interfaces.NotificationRepository,
	couponUsageRepo interfaces.CouponUsageRepository,
	notificationUseCase interfaces.NotificationUseCase,
	interval time.Duration,
	log *slog.Logger,
) *Worker {
	if interval <= 0 {
		interval = time.Minute
	}

	return &Worker{
		couponRepo:          couponRepo,
		notificationRepo:    notificationRepo,
		couponUsageRepo:     couponUsageRepo,
		notificationUseCase: notificationUseCase,
		interval:            interval,
		log:                 log,
	}
}
func (w *Worker) processExpiredCoupons(ctx context.Context) {
	now := time.Now()
	coupons, err := w.couponRepo.GetExpiredCoupons(ctx, now)
	if err != nil {
		w.log.Error("failed to fetch expired coupons", "error", err)
		return
	}
	for _, coupon := range coupons {
		eventData, err := json.Marshal(map[string]any{
			"expires_at": coupon.ExpiresAt.UTC().Format(time.RFC3339Nano),
		})
		if err != nil {
			w.log.Error(
				"failed to build coupon expiry event data",
				"coupon_id", coupon.ID,
				"error", err,
			)
			continue
		}
		exists, err := w.notificationRepo.ExistsByEventData(
			ctx,
			"coupon_expired",
			"coupon",
			coupon.ID.String(),
			eventData,
		)
		if err != nil {
			w.log.Error(
				"failed to check coupon expiry notification",
				"coupon_id", coupon.ID,
				"error", err,
			)
			continue
		}
		if exists {
			continue
		}
		entityType := "coupon"
		entityID := coupon.ID.String()
		notification := &models.Notification{
			Type:       "coupon_expired",
			Title:      "Coupon expired",
			Message:    fmt.Sprintf("Coupon %s has expired.", coupon.Code),
			Priority:   "medium",
			EntityType: &entityType,
			EntityID:   &entityID,
			EventData:  (*json.RawMessage)(&eventData),
		}

		if err := w.notificationUseCase.CreateNotification(ctx, notification); err != nil {
			w.log.Error(
				"failed to create coupon expiry notification",
				"coupon_id", coupon.ID,
				"error", err,
			)
			continue
		}

		w.log.Info(
			"coupon expiry notification created",
			"coupon_id", coupon.ID,
			"coupon_code", coupon.Code,
		)
	}
}
func (w *Worker) processExhaustedCoupons(ctx context.Context) {
	coupons, err := w.couponRepo.GetExhaustedCoupons(ctx)
	if err != nil {
		w.log.Error("failed to fetch exhausted coupons", "error", err)
		return
	}
	for _, coupon := range coupons {
		if coupon.UsageLimit == nil {
			continue
		}
		usageCount, err := w.couponUsageRepo.CountByCouponID(ctx, coupon.ID)
		if err != nil {
			w.log.Error(
				"failed to count coupon usage",
				"coupon_id", coupon.ID,
				"error", err,
			)
			continue
		}
		eventData, err := json.Marshal(map[string]any{
			"usage_limit": *coupon.UsageLimit,
		})
		if err != nil {
			w.log.Error(
				"failed to build coupon exhaustion event data",
				"coupon_id", coupon.ID,
				"error", err,
			)
			continue
		}
		exists, err := w.notificationRepo.ExistsByEventData(
			ctx,
			"coupon_exhausted",
			"coupon",
			coupon.ID.String(),
			json.RawMessage(eventData),
		)
		if err != nil {
			w.log.Error(
				"failed to check coupon exhaustion notification",
				"coupon_id", coupon.ID,
				"error", err,
			)
			continue
		}
		if exists {
			continue
		}
		entityType := "coupon"
		entityID := coupon.ID.String()
		notification := &models.Notification{
			Type:  "coupon_exhausted",
			Title: "Coupon usage limit reached",
			Message: fmt.Sprintf(
				"Coupon %s has reached its usage limit of %d uses.",
				coupon.Code,
				*coupon.UsageLimit,
			),
			Priority:   "high",
			EntityType: &entityType,
			EntityID:   &entityID,
			EventData:  (*json.RawMessage)(&eventData),
		}

		if err := w.notificationUseCase.CreateNotification(ctx, notification); err != nil {
			w.log.Error(
				"failed to create coupon exhaustion notification",
				"coupon_id", coupon.ID,
				"error", err,
			)
			continue
		}

		w.log.Info(
			"coupon exhaustion notification created",
			"coupon_id", coupon.ID,
			"coupon_code", coupon.Code,
			"usage_count", usageCount,
			"usage_limit", *coupon.UsageLimit,
		)
	}
}
func (w *Worker) processAvailableCoupons(ctx context.Context) {
	now := time.Now()

	coupons, err := w.couponRepo.GetAvailableCoupons(ctx, now)
	if err != nil {
		w.log.Error(
			"failed to fetch available coupons",
			"error", err,
		)
		return
	}

	for _, coupon := range coupons {
		if err := ctx.Err(); err != nil {
			return
		}

		eventData, err := json.Marshal(map[string]any{
			"start_at": coupon.StartAt.UTC().Format(time.RFC3339Nano),
		})
		if err != nil {
			w.log.Error(
				"failed to marshal coupon availability event data",
				"coupon_id", coupon.ID,
				"error", err,
			)
			continue
		}

		exists, err := w.notificationRepo.ExistsByEventData(
			ctx,
			"coupon_available",
			"coupon",
			coupon.ID.String(),
			json.RawMessage(eventData),
		)
		if err != nil {
			w.log.Error(
				"failed to check coupon availability notification",
				"coupon_id", coupon.ID,
				"error", err,
			)
			continue
		}

		if exists {
			continue
		}

		entityType := "coupon"
		entityID := coupon.ID.String()

		notification := &models.Notification{
			Type:       "coupon_available",
			Title:      "Coupon is now available",
			Message:    fmt.Sprintf("Coupon %s is now available for customers.", coupon.Code),
			Priority:   "medium",
			EntityType: &entityType,
			EntityID:   &entityID,
			EventData:  (*json.RawMessage)(&eventData),
		}

		if err := w.notificationUseCase.CreateNotification(ctx, notification); err != nil {
			w.log.Error(
				"failed to create coupon availability notification",
				"coupon_id", coupon.ID,
				"error", err,
			)
			continue
		}

		w.log.Info(
			"coupon availability notification created",
			"coupon_id", coupon.ID,
			"coupon_code", coupon.Code,
		)
	}
}
func (w *Worker) processCoupons(ctx context.Context) {
	w.log.Debug("starting coupon maintenance check")
	w.processExpiredCoupons(ctx)
	w.processExhaustedCoupons(ctx)
	w.processAvailableCoupons(ctx)
	w.log.Debug("finished coupon maintenance check")
}
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	w.log.Info("coupon maintenance worker started", "interval", w.interval)
	w.processCoupons(ctx)
	select {
	case <-ctx.Done():
		w.log.Info("coupon maintenance worker stopped")
		return
	case <-ticker.C:
		w.processCoupons(ctx)
	}
}
