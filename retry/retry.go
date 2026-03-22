package retry

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// Defaults for the retry loop.
const (
	defaultMaxAttempts = 3
	defaultInitDelay   = 500 * time.Millisecond
	defaultMaxDelay    = 10 * time.Second
	defaultMultiplier  = 2.0
)

// Options configures the retry behaviour.
type Options struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// Option is a functional option for configuring retry behaviour.
type Option func(*Options)

func WithMaxAttempts(n int) Option {
	return func(o *Options) { o.MaxAttempts = n }
}

func WithInitialDelay(d time.Duration) Option {
	return func(o *Options) { o.InitialDelay = d }
}

func WithMaxDelay(d time.Duration) Option {
	return func(o *Options) { o.MaxDelay = d }
}

func WithMultiplier(m float64) Option {
	return func(o *Options) { o.Multiplier = m }
}

func defaults() Options {
	return Options{
		MaxAttempts:  defaultMaxAttempts,
		InitialDelay: defaultInitDelay,
		MaxDelay:     defaultMaxDelay,
		Multiplier:   defaultMultiplier,
	}
}

// Do retries fn until it succeeds, returns a non-retryable error, or
// exhausts all attempts. Jitter of ±50 % is applied to each delay.
func Do(fn func() error, opts ...Option) error {
	cfg := defaults()
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = defaultMaxAttempts
	}

	var err error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !IsRetryable(err) {
			return err
		}

		if attempt == cfg.MaxAttempts {
			break
		}

		jittered := jitter(delay)
		time.Sleep(jittered)

		delay = nextDelay(delay, cfg.Multiplier, cfg.MaxDelay)
	}

	return err
}

// DoWithResponse retries fn, additionally treating HTTP 429 and 5xx responses
// as retryable. On a retryable status the response body is closed before the
// next attempt. On success or a non-retryable failure the response is returned
// to the caller (who is responsible for closing the body).
func DoWithResponse(fn func() (*http.Response, error), opts ...Option) (*http.Response, error) {
	cfg := defaults()
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = defaultMaxAttempts
	}

	var resp *http.Response
	var err error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		resp, err = fn()

		if err == nil && resp != nil && (resp.StatusCode == 429 || resp.StatusCode >= 500) {
			resp.Body.Close()
			err = &HTTPError{StatusCode: resp.StatusCode}
			resp = nil
		}

		if err == nil {
			return resp, nil
		}

		if !IsRetryable(err) {
			return resp, err
		}

		if attempt == cfg.MaxAttempts {
			break
		}

		jittered := jitter(delay)
		time.Sleep(jittered)

		delay = nextDelay(delay, cfg.Multiplier, cfg.MaxDelay)
	}

	return resp, err
}

// DoCtx is like Do but respects context cancellation.
// It checks the context before each retry attempt and returns early if cancelled.
func DoCtx(ctx context.Context, fn func() error, opts ...Option) error {
	cfg := defaults()
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = defaultMaxAttempts
	}

	var err error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = fn()
		if err == nil {
			return nil
		}

		if !IsRetryable(err) {
			return err
		}

		if attempt == cfg.MaxAttempts {
			break
		}

		jittered := jitter(delay)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(jittered):
		}

		delay = nextDelay(delay, cfg.Multiplier, cfg.MaxDelay)
	}

	return err
}

// DoWithResponseCtx is like DoWithResponse but respects context cancellation.
// It checks the context before each retry attempt and returns early if cancelled.
func DoWithResponseCtx(ctx context.Context, fn func() (*http.Response, error), opts ...Option) (*http.Response, error) {
	cfg := defaults()
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = defaultMaxAttempts
	}

	var resp *http.Response
	var err error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		resp, err = fn()

		if err == nil && resp != nil && (resp.StatusCode == 429 || resp.StatusCode >= 500) {
			resp.Body.Close()
			err = &HTTPError{StatusCode: resp.StatusCode}
			resp = nil
		}

		if err == nil {
			return resp, nil
		}

		if !IsRetryable(err) {
			return resp, err
		}

		if attempt == cfg.MaxAttempts {
			break
		}

		jittered := jitter(delay)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(jittered):
		}

		delay = nextDelay(delay, cfg.Multiplier, cfg.MaxDelay)
	}

	return resp, err
}

// nextDelay calculates the next backoff delay, capped at maxDelay.
func nextDelay(current time.Duration, multiplier float64, maxDelay time.Duration) time.Duration {
	next := time.Duration(float64(current) * multiplier)
	if next > maxDelay {
		return maxDelay
	}
	return next
}

// jitter adds ±50 % random jitter to a duration.
func jitter(d time.Duration) time.Duration {
	// Range: [0.5*d, 1.5*d]
	factor := 0.5 + rand.Float64() // [0.5, 1.5)
	return time.Duration(math.Round(float64(d) * factor))
}
