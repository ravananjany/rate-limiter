package lib

import (
	"context"

	"golang.org/x/time/rate"
)

type RateLimiterInterface interface {
	Allow() bool
	Wait(ctx context.Context) error
}

type rateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiterInterface(rps float64, b int) RateLimiterInterface {
	return &rateLimiter{limiter: rate.NewLimiter(rate.Limit(rps), b)}
}

func (r *rateLimiter) Allow() bool {
	return r.limiter.Allow()
}

func (r *rateLimiter) Wait(ctx context.Context) error {
	return r.limiter.Wait(ctx)
}
