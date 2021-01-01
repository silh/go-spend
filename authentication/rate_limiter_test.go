package authentication_test

import (
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"go-spend/authentication"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name        string
		rateLimiter *authentication.RedisRateLimiter
		expected    int
	}{
		{
			name: "seconds",
			rateLimiter: authentication.NewRedisRateLimiter([]authentication.Limit{
				{
					Suffix:   "s",
					Duration: time.Second,
					Amount:   4,
				},
				{
					Suffix:   "m",
					Duration: time.Minute,
					Amount:   100,
				},
				{
					Suffix:   "h",
					Duration: time.Hour,
					Amount:   100,
				},
			}, redisClient),
			expected: 4,
		},
		{
			name: "minutes",
			rateLimiter: authentication.NewRedisRateLimiter([]authentication.Limit{
				{
					Suffix:   "s",
					Duration: time.Second,
					Amount:   100,
				},
				{
					Suffix:   "m",
					Duration: time.Minute,
					Amount:   5,
				},
				{
					Suffix:   "h",
					Duration: time.Hour,
					Amount:   100,
				},
			}, redisClient),
			expected: 5,
		},
		{
			name: "hours",
			rateLimiter: authentication.NewRedisRateLimiter([]authentication.Limit{
				{
					Suffix:   "s",
					Duration: time.Second,
					Amount:   100,
				},
				{
					Suffix:   "m",
					Duration: time.Minute,
					Amount:   100,
				},
				{
					Suffix:   "h",
					Duration: time.Hour,
					Amount:   10,
				},
			}, redisClient),
			expected: 10,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clearRedis()
			limitContext := authentication.LimitContext{
				UserID: 1,
				Path:   "/balance",
			}
			executions := 99
			var results []chan bool
			f := func(c chan<- bool) {
				c <- test.rateLimiter.IsUnderLimits(limitContext)
			}
			for i := 0; i < executions; i++ {
				results = append(results, make(chan bool, 1))
				f(results[i])
			}
			numberOfTrue := 0
			for _, c := range results {
				if ok := <-c; ok {
					numberOfTrue++
				}
			}
			assert.Equal(t, test.expected, numberOfTrue)
		})
	}
}

func TestWithRedisUnavailableAlwaysTrue(t *testing.T) {
	limiter := authentication.NewRedisRateLimiter([]authentication.Limit{
		{
			Suffix:   "s",
			Duration: time.Second,
			Amount:   1,
		},
		{
			Suffix:   "m",
			Duration: time.Minute,
			Amount:   10,
		},
		{
			Suffix:   "h",
			Duration: time.Hour,
			Amount:   10,
		},
	}, redis.NewClient(&redis.Options{Addr: "localhost:9999"}))

	limitContext := authentication.LimitContext{
		UserID: 12,
		Path:   "/aa",
	}
	assert.True(t, limiter.IsUnderLimits(limitContext))
	assert.True(t, limiter.IsUnderLimits(limitContext))
	assert.True(t, limiter.IsUnderLimits(limitContext))
}
