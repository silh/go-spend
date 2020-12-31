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
	IncorrectValues         = "Incorrect values provided"
	UserOrPasswordIncorrect = "User or password incorrect"
	ServerError             = "Server Error"
	Forbidden               = "Forbidden"
	NotFound                = "Not Found"
)

var (
	ErrUserContextNotFound = errors.New("user context not found")
)

// Maps HTTP request to proper service. Validates parameters before passing them
type Router struct {
	mux http.Handler

	authenticator   authentication.Authenticator
	authorizer      authentication.Authorizer
	balanceService  expenses.BalanceService
	expensesService expenses.Service
	groupService    expenses.GroupService
	userService     authentication.UserService
}

func NewRouter(
	authenticator authentication.Authenticator,
	authorizer authentication.Authorizer,
	balanceService expenses.BalanceService,
	expensesService expenses.Service,
	groupService expenses.GroupService,
	userService authentication.UserService,
) *Router {
	mux := http.NewServeMux()
	r := &Router{
		mux:             mux,
		authenticator:   authenticator,
		authorizer:      authorizer,
		balanceService:  balanceService,
		expensesService: expensesService,
		groupService:    groupService,
		userService:     userService,
	}
	mux.Handle("/users", http.HandlerFunc(r.users))
	mux.Handle("/expenses", r.authorizer.Authorize(r.expenses))
	mux.Handle("/groups", r.authorizer.Authorize(r.groups))
	mux.Handle("/authenticate", http.HandlerFunc(r.authenticate))
	mux.Handle("/balance", r.authorizer.Authorize(r.balance))
	return r
}

// ServeHTTP delegates handling to inner http.ServeMux
func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	router.mux.ServeHTTP(writer, request)
}

// users handles all request to the /users endpoint, at the moment that only Create
func (router *Router) users(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, NotFound, http.StatusNotFound)
		return
	}
	router.createUser(w, r)
}

// createUser prepares body of request to create user and call necessary service to do the job
// If everything is correct - responds with 201
func (router *Router) createUser(w http.ResponseWriter, r *http.Request) {
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

// groups handles all requests to /groups endpoint - create and add user.
func (router *Router) groups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		router.createGroup(w, r)
	case http.MethodPut:
		router.addToGroup(w, r)
	default:
		http.Error(w, NotFound, http.StatusNotFound)
	}
}

// createGroup prepares request body and start group creation
// If everything is correct - responds with 201
func (router *Router) createGroup(w http.ResponseWriter, r *http.Request) {
	var createGroupRequest expenses.CreateGroupRequest
	var err error
	userContext, err := extractUser(r)
	if err != nil {
		http.Error(w, Forbidden, http.StatusForbidden)
		return
	}
	if err = json.NewDecoder(r.Body).Decode(&createGroupRequest); err != nil {
		http.Error(w, IncorrectBody, http.StatusBadRequest)
		return
	}
	if userContext.GroupID != 0 {
		http.Error(w, IncorrectValues, http.StatusBadRequest)
		return
	}
	createGroupContext := expenses.CreateGroupContext{
		Name:      createGroupRequest.Name,
		CreatorID: userContext.UserID,
	}
	createdGroup, err := router.groupService.Create(r.Context(), createGroupContext)
	if err != nil {
		handleGroupCreationErrors(w, err, createGroupContext)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(createdGroup); err != nil {
		http.Error(w, ServerError, http.StatusInternalServerError)
		log.Error("couldn't write response after group %s creation - %s", err)
	}
}

// authenticate performs user authentication
// If everything is correct - responds 200 and provides access and refresh tokens
func (router *Router) authenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, NotFound, http.StatusNotFound)
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
	createGroupRequest expenses.CreateGroupContext,
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

// expenses handles all request to /expenses ednpoint. At the moment that's just Create Expense
func (router *Router) expenses(w http.ResponseWriter, r *http.Request) {
	var err error
	userContext, err := extractUser(r)
	if err != nil {
		http.Error(w, Forbidden, http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, NotFound, http.StatusNotFound)
		return
	}
	router.createExpense(w, r, userContext)
}

// createExpense prepares incoming body and start expense creation.
// If everything is correct - responds with 201
func (router *Router) createExpense(
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

// addToGroup prepares incoming body and starts procedure to add user into a group
// If everything is correct - responds with 200 without a body
func (router *Router) addToGroup(w http.ResponseWriter, r *http.Request) {
	user, err := extractUser(r)
	if err != nil {
		http.Error(w, Forbidden, http.StatusForbidden)
		return
	}
	var addRequest expenses.AddToGroupRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&addRequest); err != nil {
		http.Error(w, IncorrectBody, http.StatusBadRequest)
		return
	}
	if user.GroupID != addRequest.GroupID {
		http.Error(w, IncorrectBody, http.StatusForbidden)
		return
	}
	if err = router.groupService.AddUserToGroup(r.Context(), addRequest); err != nil {
		if err == expenses.ErrUserOrGroupNotFound || err == expenses.ErrUserIsInAnotherGroup {
			http.Error(w, IncorrectValues, http.StatusBadRequest)
			return
		}
		http.Error(w, ServerError, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// balance handles request to /balance endpoint. At the moment that's only GET of a balance for a current user.
func (router *Router) balance(w http.ResponseWriter, r *http.Request) {
	var err error
	user, err := extractUser(r)
	if err != nil {
		http.Error(w, Forbidden, http.StatusForbidden)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, NotFound, http.StatusNotFound)
		return
	}
	balance, err := router.balanceService.Get(r.Context(), user.UserID)
	if err != nil {
		http.Error(w, ServerError, http.StatusInternalServerError)
		log.Error("couldn't get balance for the user - %s", err)
		return
	}
	if err = json.NewEncoder(w).Encode(&balance); err != nil {
		http.Error(w, ServerError, http.StatusInternalServerError)
		log.Error("couldn't write balance response for the user - %s", err)
	}
}

// extractUser from request. It should be put there by Authorizer
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
