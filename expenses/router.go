package expenses

import (
	"encoding/json"
	"go-spend/log"
	"net/http"
)

// Maps HTTP request to proper service. Validates parameters before passing them
type Router struct {
	userService UserService
	mux         http.Handler
}

// Create new router instance
func NewRouter(userService UserService) *Router {
	mux := http.NewServeMux()
	r := &Router{userService: userService, mux: mux}
	mux.Handle("/users", http.HandlerFunc(r.handleUsers))
	return r
}

func (router *Router) GetMux() http.Handler {
	return router.mux
}

func (router *Router) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	router.handleCreateUser(w, r)
}

func (router *Router) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var rawRequest RawCreateUserRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&rawRequest); err != nil {
		http.Error(w, "Incorrect body", http.StatusBadRequest)
		return
	}
	createUserRequest, err := ValidCreateUserRequest(rawRequest)
	if err != nil {
		http.Error(w, "Incorrect body", http.StatusBadRequest)
		return
	}
	createdUser, err := router.userService.Create(r.Context(), createUserRequest)
	if err != nil {
		if err == ErrEmailAlreadyExists {
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
			rawRequest.Email,
			err,
		)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	log.Info("Created a new user - %s", createdUser.Email)
}
