package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"go-spend/authentication"
	"go-spend/expenses"
	"go-spend/log"
	"net/http"
	"time"
)

// Config of the Application
type Config struct {
	Port                 uint
	ServerRequestTimeout uint
	DB                   DBConfig
}

// Configuration of DB connection
type DBConfig struct {
	ConnectionString string
	User             string
	Password         string
	Name             string
	SocketTimeout    time.Duration
	ConnectTimeout   time.Duration
}

// Container for all application things.
type Application struct {
	config      Config
	userService authentication.UserService
}

// Create new Application to handle expenses
func NewApplication(config Config) (*Application, error) {
	ctx := context.Background()
	if config.Port > 65536 || config.Port == 0 {
		return nil, fmt.Errorf("incorrect port value %d, should be between 1 and 65536", config.Port)
	}
	db, err := pgxpool.Connect(ctx, config.DB.ConnectionString)
	if err != nil {
		return nil, err
	}
	userRepository := expenses.NewPgUserRepository()
	userService := authentication.NewDefaultUserService(db, &authentication.BCryptPasswordEncoder{}, userRepository)

	return &Application{config: config, userService: userService}, nil
}

func (a *Application) Start() error {
	mux := http.NewServeMux()
	mux.Handle("/users", http.HandlerFunc(a.handleCreateUser))

	log.Info("Starting a server on port 8080...")
	return http.ListenAndServe(":8080", mux)
}

func (a *Application) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not supported", http.StatusBadRequest)
		return
	}
	var createUserRequest expenses.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&createUserRequest); err != nil {
		http.Error(w, "Incorrect body", http.StatusBadRequest)
		return
	}
	createdUser, err := a.userService.Create(r.Context(), createUserRequest)
	if err != nil {
		if err == expenses.ErrEmailAlreadyExists {
			http.Error(w, "User already exists", http.StatusBadRequest)
			return
		}
		log.Error("error while trying to create a user with email %s - %s", createUserRequest.Email, err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(createdUser); err != nil {
		log.Error(
			"Could not write body to the create user response with email %s - %s",
			createUserRequest.Email,
			err,
		)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	log.Info("Created a new user - %s", createdUser.Email)
}
