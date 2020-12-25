package main

import (
	"flag"
	"go-spend/expenses"
	"go-spend/log"
	"time"
)

func init() {
	log.Level = log.DebugLvl
}

func main() {
	config := expenses.Config{}
	flag.UintVar(&config.Port, "port", 8080, "Server port")
	flag.StringVar(&config.DB.Name, "db-user", "user", "DB username")
	flag.StringVar(&config.DB.Password, "db-password", "password", "DB password")
	flag.StringVar(&config.DB.Name, "db-name", "expenses", "Name of the DB")
	flag.DurationVar(&config.DB.ConnectTimeout, "db-connect-timeout", 5*time.Second, "Timeout for DB connection")
	flag.DurationVar(&config.DB.SocketTimeout, "db-socket-timeout", 20*time.Second, "Timeout for read operations")
	flag.StringVar(
		&config.DB.ConnectionString,
		"db-connection-string",
		"postgresql://locahost:5432/expenses?user=user&password=password&socketTimeout=20",
		"Connection string to access database that contains all necessary timeous",
	)
	flag.Parse()

	application, err := expenses.NewApplication(config)
	if err != nil {
		log.FatalErr(err)
	}
	err = application.Start()
	if err != nil {
		log.FatalErr(err)
	}
}
