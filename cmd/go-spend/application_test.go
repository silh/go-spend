package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/authentication"
	"go-spend/cmd/go-spend"
	"go-spend/expenses"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	pizzaPrice  = 44
	coffeePrice = 10
)

var defaultConfig = main.Config{
	ServerRequestTimeout: 20 * time.Second,
	DB: main.DBConfig{
		ConnectionString: createPGContainerAndGetDbUrl(context.Background()),
		SchemaLocation:   "../../db/001_schema.sql",
	},
	Redis: main.RedisConfig{
		Addr:     createRedisContainer(context.Background()),
		Password: redisPassword,
	},
	Security: main.SecurityConfig{
		AccessSecret:  "1234321",
		RefreshSecret: "zzzzz",
	},
}

func TestNewApplication(t *testing.T) {
	defer cleanUpDB(t, context.Background())
	port, err := getFreePort()
	require.NoError(t, err)
	config := defaultConfig
	config.Port = uint(port)

	application, err := main.NewApplication(&config)
	require.NoError(t, err)
	assert.NotNil(t, application)

	errC := make(chan error)
	go func() {
		errC <- application.Start()
	}()
	serverAddr := fmt.Sprintf("http://localhost:%d", port)
	healthCheck(t, serverAddr)

	//Check if there was an error when starting
	select {
	case err = <-errC:
		t.Error(err)
	default:
	}

	//create 4 users, 2 groups
	user1 := createUser(t, serverAddr, "1")
	user2 := createUser(t, serverAddr, "2")
	user3 := createUser(t, serverAddr, "3")
	user4 := createUser(t, serverAddr, "4")
	user1.authenticate(t)
	user4.authenticate(t)
	groupName1 := "gr1"
	groupName2 := "group2"
	group1ID := user1.createGroup(t, groupName1)
	user4.createGroup(t, groupName2)
	user1.authenticate(t) // due to current limitations need to reauthenticate after group creation
	user4.authenticate(t)
	//add users to group 1
	user1.addUserToGroup(t, user2.ID, group1ID)
	user2.authenticate(t)
	user2.addUserToGroup(t, user3.ID, group1ID)
	user3.authenticate(t)
	// start paying
	user1.payForPizza(t)
	user2.payForCoffee(t)
	checkBalances(t, user1, user2, user3, err)

	err = application.Stop()
	if err != nil && err != http.ErrServerClosed {
		t.Error(err)
	}
}

func TestApplicationFailsWithIncorrectPort(t *testing.T) {
	tests := []struct {
		name string
		port uint
		err  string
	}{
		{
			name: "0",
			port: 0,
			err:  "incorrect port value 0, should be between 1 and 65535",
		},
		{
			name: "65536",
			port: 65536,
			err:  "incorrect port value 65536, should be between 1 and 65535",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := defaultConfig
			config.Port = test.port
			application, err := main.NewApplication(&config)
			require.EqualError(t, err, test.err)
			assert.Nil(t, application)
		})
	}
}

func TestFailsWithIncorrectDBAddress(t *testing.T) {
	config := defaultConfig
	config.Port = 8080
	config.DB.ConnectionString = "uups"
	application, err := main.NewApplication(&config)
	require.Error(t, err)
	assert.Nil(t, application)
}

func TestFailsWithoutSchemaPath(t *testing.T) {
	config := defaultConfig
	config.Port = 8080
	config.DB.SchemaLocation = ""
	application, err := main.NewApplication(&config)
	require.Error(t, err)
	assert.Nil(t, application)
}

func TestFailsWithIncorrectSchemaLocation(t *testing.T) {
	config := defaultConfig
	config.Port = 8080
	config.DB.ConnectionString = "./no_schema.sql"
	application, err := main.NewApplication(&config)
	require.Error(t, err)
	assert.Nil(t, application)
}

func checkBalances(t *testing.T, user1 systemUser, user2 systemUser, user3 systemUser, err error) {
	balance1 := user1.requestBalance(t)
	balance2 := user2.requestBalance(t)
	balance3 := user3.requestBalance(t)

	user1expectedBalance := expenses.Balance{ // well it's not exactly as in description because we have uint percentages
		2: pizzaPrice*0.33 - coffeePrice*0.5,
		3: pizzaPrice * 0.34,
	}
	assert.InDelta(t, user1expectedBalance[2], balance1[2], 0.01)
	assert.InDelta(t, user1expectedBalance[3], balance1[3], 0.01)
	user2expectedBalance := expenses.Balance{
		1: -pizzaPrice*0.33 + coffeePrice*0.5,
	}
	assert.InDelta(t, user2expectedBalance[1], balance2[1], 0.01)
	assert.Zero(t, user2expectedBalance[3])
	user3expectedBalance := expenses.Balance{
		1: -pizzaPrice * 0.34,
	}
	require.NoError(t, err)
	assert.InDelta(t, user3expectedBalance[1], balance3[1], 0.01)
}

func healthCheck(t *testing.T, serverAddr string) {
	finishPolling := make(chan struct{})
	healthC := make(chan struct{})
	go func() { // poll for 200 OK
		for {
			select {
			case <-finishPolling:
				return
			default:
				response, err := http.Get(serverAddr + "/health")
				if err == nil && response.StatusCode == http.StatusOK {
					close(healthC)
					return
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
	// wait for healthcheck
	select {
	case <-time.After(3 * time.Second):
		t.Error("didn't get 200 OK in time")
		close(finishPolling)
	case <-healthC:
	}
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func createUser(t *testing.T, serverAddr string, emailPrefix string) systemUser {
	email := emailPrefix + "mail@mail.com"
	password := emailPrefix + "123621"
	body := fmt.Sprintf(`{"email":"%s", "password":"%s"}`, email, password)
	result, err := http.Post(serverAddr+"/users", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer result.Body.Close()
	require.Equal(t, http.StatusCreated, result.StatusCode)
	var response expenses.UserResponse
	require.NoError(t, json.NewDecoder(result.Body).Decode(&response))
	return systemUser{
		ID:         response.ID,
		Email:      email,
		Password:   password,
		serverAddr: serverAddr,
	}
}

type systemUser struct {
	ID           uint
	Password     string
	Email        string
	AccessToken  string // it is empty at the start, didn't want to create different struct in tests
	RefreshToken string // it is empty at the start, didn't want to create different struct in tests
	GroupID      uint   // it is empty at the start, didn't want to create different struct in tests
	serverAddr   string
}

func (u *systemUser) authenticate(t *testing.T) {
	body := fmt.Sprintf(`{"email":"%s", "password":"%s"}`, u.Email, u.Password)
	result, err := http.Post(u.serverAddr+"/authenticate", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer result.Body.Close()
	require.Equal(t, http.StatusOK, result.StatusCode)
	var auth authentication.TokenResponse
	require.NoError(t, json.NewDecoder(result.Body).Decode(&auth))
	u.AccessToken = auth.AccessToken
	u.RefreshToken = auth.RefreshToken
}

func (u *systemUser) createGroup(t *testing.T, groupName string) uint {
	body := fmt.Sprintf(`{"name":"%s"}`, groupName)
	request, err := http.NewRequest(http.MethodPost, u.serverAddr+"/groups", strings.NewReader(body))
	require.NoError(t, err)
	u.addAuthHeader(request)
	result, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer result.Body.Close()
	require.Equal(t, http.StatusCreated, result.StatusCode)
	var response expenses.GroupResponse
	require.NoError(t, json.NewDecoder(result.Body).Decode(&response))
	return response.ID
}

func (u *systemUser) addUserToGroup(t *testing.T, userID uint, groupID uint) {
	body := fmt.Sprintf(`{"userId": %d, "groupId": %d}`, userID, groupID)
	request, err := http.NewRequest(http.MethodPut, u.serverAddr+"/groups", strings.NewReader(body))
	u.addAuthHeader(request)
	require.NoError(t, err)
	result, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer result.Body.Close()
	require.Equal(t, http.StatusOK, result.StatusCode)
}

func (u *systemUser) payForPizza(t *testing.T) {
	body := `
	{
		"amount": 44,
		"shares": {
			"1": 33,
			"2": 33,
			"3": 34
		}
	}`
	u.payForExpense(t, body)
}

func (u *systemUser) payForCoffee(t *testing.T) {
	body := `
	{
		"amount": 10,
		"shares": {
			"1": 50,
			"2": 50
		}
	}`
	u.payForExpense(t, body)
}

func (u *systemUser) payForExpense(t *testing.T, expenseBody string) {
	request, err := http.NewRequest(http.MethodPost, u.serverAddr+"/expenses", strings.NewReader(expenseBody))
	u.addAuthHeader(request)
	require.NoError(t, err)
	result, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer result.Body.Close()
	require.Equal(t, http.StatusCreated, result.StatusCode)
}

func (u *systemUser) requestBalance(t *testing.T) expenses.Balance {
	request, err := http.NewRequest(http.MethodGet, u.serverAddr+"/balance", nil)
	u.addAuthHeader(request)
	require.NoError(t, err)
	result, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer result.Body.Close()
	require.Equal(t, http.StatusOK, result.StatusCode)
	var balance expenses.Balance
	require.NoError(t, json.NewDecoder(result.Body).Decode(&balance))
	return balance
}

func (u *systemUser) addAuthHeader(r *http.Request) {
	if r == nil {
		return
	}
	r.Header.Add("Authorization", "Bearer "+u.AccessToken)
}
