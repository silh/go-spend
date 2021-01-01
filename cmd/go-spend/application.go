package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jackc/pgx/v4/pgxpool"
	"go-spend/authentication"
	"go-spend/authentication/jwt"
	"go-spend/expenses"
	"go-spend/log"
	"io/ioutil"
	"net/http"
	"time"
)

// Config of the Application
type Config struct {
	Port                 uint
	ServerRequestTimeout time.Duration
	DB                   DBConfig
	Redis                RedisConfig
	Security             SecurityConfig
}

// DBConfig contains information about DB connectivity
type DBConfig struct {
	ConnectionString string
	SchemaLocation   string
}

// RedisConfig contains properties for redis connection
type RedisConfig struct {
	Addr     string
	Password string
}

// SecurityConfig contains keys for generated tokens
type SecurityConfig struct {
	AccessSecret  string
	RefreshSecret string
}

// Application constructs all parts and starts the work of the system
type Application struct {
	server *http.Server
}

// NewApplication does all necessary preparations to start the application server
func NewApplication(config *Config) (*Application, error) {
	ctx := context.Background()
	if config.Port < 1 || config.Port > 65535 {
		return nil, fmt.Errorf("incorrect port value %d, should be between 1 and 65535", config.Port)
	}
	db, err := prepareDB(ctx, config)
	if err != nil {
		return nil, err
	}
	accessAlg := jwt.HmacSha256(config.Security.AccessSecret)
	refreshAlg := jwt.HmacSha256(config.Security.RefreshSecret)
	tokenCreator := authentication.NewTokenCreator(accessAlg, refreshAlg)
	redisClient := redis.NewClient(&redis.Options{Addr: config.Redis.Addr, Password: config.Redis.Password})
	tokenRepository := authentication.NewRedisTokenRepository(redisClient)
	passwordEncoder := authentication.NewBCryptPasswordEncoder()
	userRepository := expenses.NewPgUserRepository()
	authService := authentication.NewAuthService(db, tokenCreator, tokenRepository, passwordEncoder, userRepository)

	authorizer := authentication.NewJWTAuthorizer(accessAlg, tokenRepository)

	repository := expenses.NewPgBalanceRepository()
	balanceService := expenses.NewDefaultBalanceService(db, repository)

	groupRepository := expenses.NewPgGroupRepository()
	expensesRepository := expenses.NewPgRepository()
	expensesServices := expenses.NewDefaultService(db, groupRepository, expensesRepository)

	groupService := expenses.NewDefaultGroupService(db, userRepository, groupRepository)

	userService := authentication.NewDefaultUserService(db, &authentication.BCryptPasswordEncoder{}, userRepository)

	router := NewRouter(authService, authorizer, balanceService, expensesServices, groupService, userService)
	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", config.Port),
		Handler:     router,
		ReadTimeout: config.ServerRequestTimeout,
	}
	return &Application{server: server}, nil
}

// Start a server and block until finished
func (a *Application) Start() error {
	log.Info("Starting a server on %s...", a.server.Addr)
	return a.server.ListenAndServe()
}

func (a *Application) Stop() error {
	log.Info("Stopping the server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.server.Shutdown(ctx)
}

func prepareDB(ctx context.Context, config *Config) (*pgxpool.Pool, error) {
	if config.DB.SchemaLocation == "" {
		return nil, errors.New("schema location is not specified")
	}
	db, err := pgxpool.Connect(ctx, config.DB.ConnectionString)
	schema, err := ioutil.ReadFile(config.DB.SchemaLocation)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(ctx, string(schema))
	if err != nil {
		return nil, err
	}
	return db, nil
}
