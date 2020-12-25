package expenses

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"go-spend/log"
	"net/http"
	"time"
)

// Config of the Application
type Config struct {
	Port uint
	DB   DBConfig
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
	config     Config
	repository UserRepository
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
	userRepository := NewPgRepository(db)

	return &Application{config: config, repository: userRepository}, nil
}

func (a *Application) Start() error {
	mux := http.NewServeMux()
	mux.Handle("/users", http.HandlerFunc(handleCreateUser))

	log.Info("Starting a server on port 8080...")
	return http.ListenAndServe(":8080", mux)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not supported", http.StatusBadRequest)
		return
	}
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Incorrect body", http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusCreated)

}
