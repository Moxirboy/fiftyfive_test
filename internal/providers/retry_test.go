package providers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"flysoft-flight-service/internal/providers"
)

var errBoom = errors.New("provider boom")

type scriptedResult struct {
	offers []providers.ProviderOffer
	err    error
	delay  time.Duration
}

type scriptedProvider struct {
	name    string
	calls   int
	results []scriptedResult
}

func (p *scriptedProvider) Name() string { return p.name }

func (p *scriptedProvider) Search(ctx context.Context, _ providers.ProviderSearch) ([]providers.ProviderOffer, error) {
	idx := p.calls
	p.calls++
	if idx >= len(p.results) {
		idx = len(p.results) - 1
	}
	res := p.results[idx]

	if res.delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(res.delay):
		}
	}

	return res.offers, res.err
}

func sampleOffers() []providers.ProviderOffer {
	return []providers.ProviderOffer{{ProviderRef: "SUP-001", From: "DYU", To: "IST"}}
}

func TestRetryProviderSucceedsFirstAttempt(t *testing.T) {
	provider := &scriptedProvider{name: "Scripted", results: []scriptedResult{{offers: sampleOffers()}}}
	wrapped := providers.NewRetry(provider, providers.RetryConfig{MaxRetries: 2})

	offers, err := wrapped.Search(context.Background(), providers.ProviderSearch{})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(offers) != 1 {
		t.Fatalf("offers = %d, want 1", len(offers))
	}
	if provider.calls != 1 {
		t.Fatalf("calls = %d, want 1", provider.calls)
	}
	if wrapped.Name() != "Scripted" {
		t.Fatalf("Name() = %q, want Scripted", wrapped.Name())
	}
}

func TestRetryProviderRetriesThenSucceeds(t *testing.T) {
	provider := &scriptedProvider{
		name: "Scripted",
		results: []scriptedResult{
			{err: errBoom},
			{err: errBoom},
			{offers: sampleOffers()},
		},
	}
	wrapped := providers.NewRetry(provider, providers.RetryConfig{MaxRetries: 2})

	offers, err := wrapped.Search(context.Background(), providers.ProviderSearch{})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(offers) != 1 {
		t.Fatalf("offers = %d, want 1", len(offers))
	}
	if provider.calls != 3 {
		t.Fatalf("calls = %d, want 3", provider.calls)
	}
}

func TestRetryProviderExhaustsRetries(t *testing.T) {
	provider := &scriptedProvider{name: "Scripted", results: []scriptedResult{{err: errBoom}}}
	wrapped := providers.NewRetry(provider, providers.RetryConfig{MaxRetries: 2})

	_, err := wrapped.Search(context.Background(), providers.ProviderSearch{})
	if !errors.Is(err, errBoom) {
		t.Fatalf("Search() error = %v, want %v", err, errBoom)
	}
	if provider.calls != 3 {
		t.Fatalf("calls = %d, want 3 (1 + 2 retries)", provider.calls)
	}
}

func TestRetryProviderTimesOutPerAttemptThenRetries(t *testing.T) {
	provider := &scriptedProvider{
		name: "Scripted",
		results: []scriptedResult{
			{delay: 100 * time.Millisecond}, // exceeds per-attempt timeout
			{offers: sampleOffers()},         // fast success on retry
		},
	}
	wrapped := providers.NewRetry(provider, providers.RetryConfig{
		Timeout:    10 * time.Millisecond,
		MaxRetries: 1,
	})

	offers, err := wrapped.Search(context.Background(), providers.ProviderSearch{})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(offers) != 1 {
		t.Fatalf("offers = %d, want 1", len(offers))
	}
	if provider.calls != 2 {
		t.Fatalf("calls = %d, want 2", provider.calls)
	}
}

func TestRetryProviderStopsOnCanceledParentContext(t *testing.T) {
	provider := &scriptedProvider{name: "Scripted", results: []scriptedResult{{err: errBoom}}}
	wrapped := providers.NewRetry(provider, providers.RetryConfig{MaxRetries: 3})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := wrapped.Search(ctx, providers.ProviderSearch{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Search() error = %v, want %v", err, context.Canceled)
	}
	if provider.calls != 0 {
		t.Fatalf("calls = %d, want 0", provider.calls)
	}
}
