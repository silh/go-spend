package main

import (
	"flag"
	"go-spend/log"
	"net/http"
	"time"
)

func init() {
	log.Level = log.DebugLvl
}

func main() {
	application, err := NewApplication(PrepareConfig())
	if err != nil {
		log.FatalErr(err)
	}
	if err = application.Start(); err != nil && err != http.ErrServerClosed {
		log.FatalErr(err)
	}
}

// PrepareConfig creates app config from flags. As a possible improvement read of valeus from ENV could be introduces
// or config files.
func PrepareConfig() *Config {
	config := &Config{}
	flag.UintVar(&config.Port, "port", 8080, "Server port")
	flag.DurationVar(
		&config.ServerRequestTimeout,
		"server-request-timeout",
		20*time.Second,
		"Timeout for all server requests",
	)
	flag.StringVar(
		&config.DB.ConnectionString,
		"db-connection-string",
		"postgresql://locahost:5432/expenses?user=user&password=password&socketTimeout=20",
		"Connection string to access database",
	)
	flag.StringVar(
		&config.DB.SchemaLocation,
		"db-schema-location",
		"./001_schema.sql",
		"Location of a schema file for DB.",
	)
	flag.StringVar(
		&config.Redis.Addr,
		"redis-address",
		"localhost:6379",
		"Redis address in format host:port",
	)
	flag.StringVar(
		&config.Redis.Password,
		"redis-password",
		"",
		"Redis password. Might be empty",
	)
	flag.StringVar(
		&config.Security.AccessSecret,
		"access-token-secret",
		"access-secret",
		"Secret key for access token encryption",
	)
	flag.StringVar(
		&config.Security.RefreshSecret,
		"refresh-token-secret",
		"refresh-secret",
		"Secret key for refresh token encryption. They are not implemented at the moment",
	)
	flag.Parse()
	return config
}
