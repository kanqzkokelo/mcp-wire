package auth

import (
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(reqPerSec float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(reqPerSec), burst),
	}
}

func (r *RateLimiter) Allow() bool {
	return r.limiter.Allow()
}
