package loadtest

import (
	"context"
	"time"

	"github.com/Ygohr/fc-stress-tester/internal/contract"
)

type Service struct {
	client contract.HTTPClient
}

func NewService(client contract.HTTPClient) *Service {
	return &Service{client: client}
}

func (s *Service) Execute(ctx context.Context, cfg Config) (Report, error) {
	pool := newWorkerPool(s.client, cfg.Concurrency)

	start := time.Now()
	results := pool.run(ctx, cfg.URL, cfg.Requests)
	duration := time.Since(start)

	return buildReport(results, duration), nil
}
