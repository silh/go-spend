package main

import (
	"flag"
	"go-spend/log"
	"net/http"
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

func PrepareConfig() Config {
	config := Config{}
	flag.UintVar(&config.Port, "port", 8080, "Server port")
	flag.StringVar(
		&config.DB.ConnectionString,
		"db-connection-string",
		"postgresql://locahost:5432/expenses?user=user&password=password&socketTimeout=20",
		"Connection string to access database",
	)
	flag.Parse()
	return config
}
