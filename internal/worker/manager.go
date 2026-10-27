package worker

import (
	"context"
	"ecommerce/internal/usecase/interfaces"
	"log/slog"
	"sync"
)

type Manager struct {
	workers []interfaces.Worker
	log     *slog.Logger
	wg      sync.WaitGroup
}

func NewManager(log *slog.Logger) *Manager {
	return &Manager{
		workers: make([]interfaces.Worker, 0),
		log:     log,
	}
}

func (m *Manager) Register(w interfaces.Worker) {
	m.workers = append(m.workers, w)
}

func (m *Manager) Start(ctx context.Context) {
	for _, w := range m.workers {
		m.wg.Add(1)

		go func(w interfaces.Worker) {
			defer m.wg.Done()
			w.Run(ctx)
		}(w)
	}

	m.log.Info("background workers started",
		"count", len(m.workers),
	)
}

func (m *Manager) Wait() {
	m.wg.Wait()
	m.log.Info("all background workers stopped")
}
