package authentication_test

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	redisImage    = "redis:6.0.9-alpine3.12"
	redisPassword = "password"
)

var (
	redisClient = redis.NewClient(&redis.Options{
		Password: redisPassword,
		Addr:     createRedisContainer(context.Background()),
	})
)

// Creates Redis container and return its address
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
