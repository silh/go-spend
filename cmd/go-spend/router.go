package main

import (
	"encoding/json"
	"errors"
	"go-spend/authentication"
	"go-spend/expenses"
	"go-spend/log"
	"net/http"
)

const (
	IncorrectBody           = "Incorrect body"
	UserOrPasswordIncorrect = "User or password incorrect"
	ServerError             = "Server Error"
	Forbidden               = "Forbidden"
)

var (
	ErrUserContextNotFound = errors.New("user context not found")
)

// Maps HTTP request to proper service. Validates parameters before passing them
type Router struct {
	mux http.Handler

	authenticator   authentication.Authenticator
	authorizer      authentication.Authorizer
	expensesService expenses.Service
	groupService    expenses.GroupService
	userService     authentication.UserService
}

func NewRouter(
	authenticator authentication.Authenticator,
	authorizer authentication.Authorizer,
	expensesService expenses.Service,
	groupService expenses.GroupService,
	userService authentication.UserService,
) *Router {
	mux := http.NewServeMux()
	r := &Router{
		mux:             mux,
		authenticator:   authenticator,
		expensesService: expensesService,
		authorizer:      authorizer,
		groupService:    groupService,
		userService:     userService,
	}
	mux.Handle("/users", http.HandlerFunc(r.handleUsers))
	mux.Handle("/expenses", r.authorizer.Authorize(r.handleExpenses))
	mux.Handle("/groups", r.authorizer.Authorize(r.handleGroups))
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
	var err error
	if err = json.NewDecoder(r.Body).Decode(&createGroupRequest); err != nil {
		http.Error(w, IncorrectBody, http.StatusBadRequest)
		return
	}
	createdGroup, err := router.groupService.Create(r.Context(), createGroupRequest)
	if err != nil {
		handleGroupCreationErrors(w, err, createGroupRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(createdGroup); err != nil {
		http.Error(w, ServerError, http.StatusInternalServerError)
		log.Error("couldn't write response after group %s creation - %s", err)
	}
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
		log.Error("couldn't write body of auth response - %s", err)
		http.Error(w, ServerError, http.StatusInternalServerError)
	}
}

func handleGroupCreationErrors(
	w http.ResponseWriter,
	err error,
	createGroupRequest expenses.CreateGroupRequest,
) {
	switch err {
	case expenses.ErrUserNotFound:
		http.Error(w, "User doesn't exists", http.StatusBadRequest)
	case expenses.ErrGroupNameAlreadyExists:
		http.Error(w, "Group with such name already exists", http.StatusBadRequest)
	case expenses.ErrUserIsInAnotherGroup:
		http.Error(w, "User participates in another group", http.StatusBadRequest)
	default:
		http.Error(w, ServerError, http.StatusInternalServerError)
		log.Error("couldn't create group %s by user %d - %s", createGroupRequest.Name, createGroupRequest.CreatorID, err)
	}
}

func (router *Router) handleExpenses(w http.ResponseWriter, r *http.Request) {
	var err error
	userContext, err := extractUser(r)
	if err != nil {
		http.Error(w, Forbidden, http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	router.handleCreateExpense(w, r, userContext)
}

func (router *Router) handleCreateExpense(
	w http.ResponseWriter,
	r *http.Request,
	userContext authentication.UserContext,
) {
	var err error
	var expenseReq expenses.CreateExpenseRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&expenseReq); err != nil {
		http.Error(w, IncorrectBody, http.StatusBadRequest)
		return
	}
	expenseContext := expenses.CreateExpenseContext{
		UserID:  userContext.UserID,
		GroupID: userContext.GroupID,
		Amount:  expenseReq.Amount,
		Shares:  expenseReq.Shares,
	}
	if err = expenses.ValidateCreateExpenseContext(expenseContext); err != nil {
		http.Error(w, IncorrectBody, http.StatusBadRequest)
		return
	}
	created, err := router.expensesService.Create(r.Context(), expenseContext)
	if err != nil {
		http.Error(w, ServerError, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(&created); err != nil {
		http.Error(w, ServerError, http.StatusInternalServerError)
		log.Error("couldn't write body for create expense response - %s", err)
	}
}

func extractUser(r *http.Request) (authentication.UserContext, error) {
	value := r.Context().Value("user")
	if value == nil {
		return authentication.UserContext{}, ErrUserContextNotFound
	}
	userContext, ok := value.(authentication.UserContext)
	if !ok {
		return authentication.UserContext{}, ErrUserContextNotFound
	}
	return userContext, nil
}
