package main

import (
	"encoding/json"
	"go-spend/authentication"
	"go-spend/expenses"
	"go-spend/log"
	"net/http"
)

const (
	IncorrectBody           = "Incorrect body"
	UserOrPasswordIncorrect = "User or password incorrect"
	ServerError             = "Server Error"
)

// Maps HTTP request to proper service. Validates parameters before passing them
type Router struct {
	userService   expenses.UserService
	groupService  expenses.GroupService
	authenticator authentication.Authenticator
	mux           http.Handler
}

// Create new router instance
func NewRouter(
	userService expenses.UserService,
	groupService expenses.GroupService,
	authenticator authentication.Authenticator) *Router {
	mux := http.NewServeMux()
	r := &Router{userService: userService, groupService: groupService, authenticator: authenticator, mux: mux}
	mux.Handle("/users", http.HandlerFunc(r.handleUsers))
	mux.Handle("/groups", http.HandlerFunc(r.handleGroups))
	mux.Handle("/authenticate", http.HandlerFunc(r.handleAuthentication))
	return r
}

// ServeHTTP delegates handling to inner http.ServeMux
func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	router.mux.ServeHTTP(writer, request)
}

func (router *Router) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	router.handleCreateUser(w, r)
}

func (router *Router) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var createUserRequest expenses.CreateUserRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&createUserRequest); err != nil {
		http.Error(w, IncorrectBody, http.StatusBadRequest)
		return
	}
	createdUser, err := router.userService.Create(r.Context(), createUserRequest)
	if err == expenses.ErrEmailAlreadyExists {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Error("error while trying to create a user with email %s - %s", createUserRequest.Email, err)
		http.Error(w, ServerError, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(createdUser); err != nil {
		log.Error("could not write body of the create user response - %s", err)
		http.Error(w, ServerError, http.StatusInternalServerError)
		return
	}
	log.Info("created a new user - %s", createdUser.Email)
}

func (router *Router) handleGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	router.handleCreateGroup(w, r)
}

func (router *Router) handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	var createGroupRequest expenses.CreateGroupRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&createGroupRequest); err != nil {
		http.Error(w, IncorrectBody, http.StatusBadRequest)
		return
	}
	//create, err := router.groupService.Create(r.Context(), createGroupRequest)
}

func (router *Router) handleAuthentication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	var auth authentication.AuthRequest
	var err error
	if err = json.NewDecoder(r.Body).Decode(&auth); err != nil {
		http.Error(w, IncorrectBody, http.StatusBadRequest)
		return
	}
	tokenResponse, err := router.authenticator.Authenticate(r.Context(), auth.Email, auth.Password)
	if err == authentication.ErrEmailOrPasswordIncorrect {
		http.Error(w, UserOrPasswordIncorrect, http.StatusUnauthorized)
		log.Info("incorrect authentication attempt %s", auth.Email)
		return
	}
	if err != nil {
		http.Error(w, ServerError, http.StatusInternalServerError)
		log.Error("could not authenticate user with email %s, error: %s", auth.Email, err)
		return
	}
	if err = json.NewEncoder(w).Encode(&tokenResponse); err != nil {
		log.Error("could not write body of auth response - %s", err)
		http.Error(w, ServerError, http.StatusInternalServerError)
	}
}
