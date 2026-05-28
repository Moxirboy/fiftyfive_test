package providers

import (
	"context"
	"time"
)

// RetryConfig controls the resilience behavior applied by a retrying
// FlightProvider wrapper.
type RetryConfig struct {
	// Timeout bounds a single provider attempt. Zero disables the per-attempt timeout.
	Timeout time.Duration
	// MaxRetries is the number of additional attempts after the first. Negative is treated as zero.
	MaxRetries int
	// Backoff is the fixed delay inserted before each retry. Zero means retry immediately.
	Backoff time.Duration
}

type retryProvider struct {
	next  FlightProvider
	cfg   RetryConfig
	sleep func(context.Context, time.Duration) error
}

// NewRetry wraps a FlightProvider so that each Search applies a per-attempt
// timeout and retries failed attempts with a fixed backoff. The caller's
// context is always honored: once it is cancelled, no further attempts run.
//
// Adding resilience here keeps FlightService and the concrete providers free of
// timeout/retry bookkeeping, and the wrapper composes with any FlightProvider.
func NewRetry(next FlightProvider, cfg RetryConfig) FlightProvider {
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}

	return &retryProvider{
		next:  next,
		cfg:   cfg,
		sleep: sleepWithContext,
	}
}

func (p *retryProvider) Name() string {
	return p.next.Name()
}

func (p *retryProvider) Search(ctx context.Context, req ProviderSearch) ([]ProviderOffer, error) {
	var lastErr error

	for attempt := 0; attempt <= p.cfg.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if attempt > 0 && p.cfg.Backoff > 0 {
			if err := p.sleep(ctx, p.cfg.Backoff); err != nil {
				return nil, err
			}
		}

		offers, err := p.searchOnce(ctx, req)
		if err == nil {
			return offers, nil
		}
		lastErr = err

		// A per-attempt timeout is a transient failure we retry; only stop when
		// the caller's own context is done.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	return nil, lastErr
}

func (p *retryProvider) searchOnce(ctx context.Context, req ProviderSearch) ([]ProviderOffer, error) {
	if p.cfg.Timeout <= 0 {
		return p.next.Search(ctx, req)
	}

	attemptCtx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	return p.next.Search(attemptCtx, req)
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
