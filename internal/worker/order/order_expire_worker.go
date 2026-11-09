package orderworker

import (
	"context"
	"log/slog"
	"time"

	"ecommerce/internal/usecase/interfaces"
)

type Worker struct {
	orderRepo    interfaces.OrderRepository
	orderUseCase interfaces.OrderUseCase

	interval time.Duration
	log      *slog.Logger
}

func NewWorker(
	orderRepo interfaces.OrderRepository,
	orderUseCase interfaces.OrderUseCase,
	interval time.Duration,
	log *slog.Logger,
) *Worker {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	return &Worker{
		orderRepo:    orderRepo,
		orderUseCase: orderUseCase,
		interval:     interval,
		log:          log,
	}
}

func (w *Worker) processExpiredOrders(ctx context.Context) {
	now := time.Now()
	orders, err := w.orderRepo.ListExpiredPendingOrders(ctx, now, 100)
	if err != nil {
		w.log.ErrorContext(
			ctx,
			"failed to fetch expired pending orders",
			"error", err,
		)
		return
	}
	for _, order := range orders {
		if ctx.Err() != nil {
			return
		}
		err := w.orderUseCase.ExpirePendingOrder(ctx, order.ID)
		if err != nil {
			w.log.ErrorContext(
				ctx,
				"failed to expire order",
				"order_id", order.ID,
				"error", err,
			)
			continue
		}
		w.log.InfoContext(
			ctx,
			"order expiry processed",
			"order_id", order.ID,
		)
	}
}
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	w.log.Info(
		"order expiry worker started",
		"interval", w.interval,
	)
	w.processExpiredOrders(ctx)
	for {
		select {
		case <-ctx.Done():
			w.log.Info("order expiry worker stopped")
			return

		case <-ticker.C:
			w.processExpiredOrders(ctx)
		}
	}
}

var _ interfaces.Worker = (*Worker)(nil)
