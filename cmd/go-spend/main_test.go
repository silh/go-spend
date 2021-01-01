package main_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/cmd/go-spend"
	"testing"
	"time"
)

var defaultConfigFromFlags = &main.Config{
	Port:                 8080,
	ServerRequestTimeout: 20 * time.Second,
	DB: main.DBConfig{
		ConnectionString: "postgresql://locahost:5432/expenses?user=user&password=password",
		SchemaLocation:   "./001_schema.sql",
	},
	Redis: main.RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
	},
	Security: main.SecurityConfig{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
	},
}

func TestPrepareConfig(t *testing.T) {
	config := main.PrepareConfig()
	require.NotNil(t, config)
	assert.Equal(t, defaultConfigFromFlags, config)
}
