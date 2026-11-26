package ratelimiter

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type RateLimiter struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64
	lastRefillTime time.Time
	mutex          sync.Mutex
}

func NewRateLimiter(maxTokens, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}
func (r *RateLimiter) refillTokens() {
	now := time.Now()
	duration := now.Sub(r.lastRefillTime).Seconds()
	tokensAdd := duration * r.refillRate
	r.tokens += tokensAdd
	if r.tokens > r.maxTokens {
		r.tokens = r.maxTokens
	}
	r.lastRefillTime = now
}

func (r *RateLimiter) IsAllow() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.refillTokens()
	if r.tokens >= 1 {
		r.tokens--
		return true
	}
	return false
}

type IpRateLimiter struct {
	limiters   *expirable.LRU[string, *RateLimiter]
	maxTokens  float64
	refillRate float64
}

func NewIpRateLimiter(ttl int, maxTokens, refillRate float64) *IpRateLimiter {
	limiter := expirable.NewLRU[string, *RateLimiter](0, nil, time.Duration(ttl)*time.Minute)
	return &IpRateLimiter{
		limiters:   limiter,
		maxTokens:  maxTokens,
		refillRate: refillRate,
	}
}

func (ir *IpRateLimiter) GetLimiter(ip string) *RateLimiter {
	_, ok := ir.limiters.Get(ip)
	if !ok {
		ir.limiters.Add(ip, NewRateLimiter(ir.maxTokens, ir.refillRate))
	}
	limiter, _ := ir.limiters.Get(ip)
	return limiter
}
