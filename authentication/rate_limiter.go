package authentication

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"go-spend/log"
	"go-spend/util"
	"net/http"
	"time"
)

const (
	justValue = "1"
)

var (
	ErrTooManyRequests = errors.New("too many requests from the same user")
)

// RateLimiter checks if limit for a particular context was reached
type RateLimiter interface {
	IsUnderLimits(context LimitContext) bool
}

// RedisRateLimiter is a RateLimiter that uses redis as a storage
type RedisRateLimiter struct {
	limits []Limit
	redis  redis.UniversalClient
}

// NewRedisRateLimiter creates new instance of RedisRateLimiter. User must ensure that passed limits are correct and
// their suffixes are unique.
func NewRedisRateLimiter(limits []Limit, redis redis.UniversalClient) *RedisRateLimiter {
	return &RedisRateLimiter{limits: limits, redis: redis}
}

// IsUnderLimits checks if limit for a context was not reached for seconds/minutes/hours.
// It is a tolerant limiter and will return true when something went wrong while trying to access redis. Only returns
// false when was able to check that the limit was reached.
func (r *RedisRateLimiter) IsUnderLimits(context LimitContext) bool {
	err := r.redis.Watch(func(tx *redis.Tx) error {
		for _, limit := range r.limits {
			key := context.AsKey(string(limit.Suffix))
			if err := checkLimitForKey(tx, context, limit); err != nil {
				if err == ErrTooManyRequests {
					return err
				}
				log.Warn("failed to check %s key - %s", key, err)
			}
		}
		return nil
	})
	return err != ErrTooManyRequests
}

func checkLimitForKey(tx *redis.Tx, limitContext LimitContext, limit Limit) error {
	key := limitContext.AsKey(string(limit.Suffix))
	curLen, err := tx.LLen(key).Result()
	if err != nil {
		return err
	}
	if uint64(curLen) >= limit.Amount {
		return ErrTooManyRequests
	}
	keyExists, err := tx.Exists(key).Result()
	if err != nil {
		return err
	}
	if keyExists == 1 {
		return tx.RPushX(key, justValue).Err()
	}
	if err := tx.RPush(key, justValue).Err(); err != nil {
		return err
	}
	return tx.Expire(key, limit.Duration).Err()
}

// LimitContext contains info to prepare base for the key in cache
type LimitContext struct {
	UserID uint
	Path   string
}

// AsKey transforms LimitContext to a value to be stored as key in cache
func (l *LimitContext) AsKey(suffix string) string {
	return fmt.Sprintf("%d_%s_%s", l.UserID, l.Path, suffix)
}

// Limit contains info for one limit. Suffix will be appended to a key created from LimitContext
type Limit struct {
	Suffix   util.NonEmptyString
	Duration time.Duration
	Amount   uint64
}

// RequestLimiter creates limited http.Handler wrapping the provided original handler.
type RequestLimiter interface {
	// RateLimit the passed handler
	RateLimit(next http.HandlerFunc) http.HandlerFunc
}

// ContextBasedRequestLimiter is RequestLimiter that extracts user info from request context and checks limit based on
// that information. If user is not present in request the request is not limited! A warning message will appear in logs
type ContextBasedRequestLimiter struct {
	rateLimiter RateLimiter
}

// NewContextBasedRequestLimiter creates new instance of ContextBasedRequestLimiter
func NewContextBasedRequestLimiter(rateLimiter RateLimiter) *ContextBasedRequestLimiter {
	return &ContextBasedRequestLimiter{rateLimiter: rateLimiter}
}

// RateLimit extract user from request context and rate limit the request
func (rh *ContextBasedRequestLimiter) RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rh.isUnderLimit(r) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
		}
		next.ServeHTTP(w, r)
	}
}

func (rh *ContextBasedRequestLimiter) isUnderLimit(r *http.Request) bool {
	user, err := ExtractUser(r)
	path := r.URL.Path
	if err != nil {
		log.Error(
			"couldn't rate limit request to %s because use not present in context, possible misconfiguration!",
			path,
		)
		return true
	}
	return rh.rateLimiter.IsUnderLimits(LimitContext{UserID: user.UserID, Path: path})
}
