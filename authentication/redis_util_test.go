package authentication_test

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const redisImage = "redis:6.0.9-alpine3.12"

var (
	redisPassword = "password"
	redisClient   = redis.NewClient(&redis.Options{
		Password: redisPassword,
		Addr:     createRedisContainer(context.Background()),
	})
)

// Creates PG container, applies necessary schema. If there is any error - it will panic
func createRedisContainer(ctx context.Context) string {
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        redisImage,
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections"),
			Cmd:          []string{"redis-server", "--requirepass", redisPassword},
		},
		Started: true,
	})
	if err != nil {
		panic(err)
	}
	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		panic(err)
	}
	return endpoint
}
