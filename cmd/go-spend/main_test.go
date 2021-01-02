package main_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go-spend/cmd/go-spend"
	"testing"
	"time"
)

var defaultConfigFromFlags = &main.Config{
	Port:                 8080,
	ServerRequestTimeout: 20 * time.Second,
	DB: main.DBConfig{
		ConnectionString: "postgresql://locahost:5432/expenses?user=user&password=password",
		SchemaLocation:   "./001_schema.sql",
	},
	Redis: main.RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
	},
	Security: main.SecurityConfig{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
	},
}

func TestPrepareConfig(t *testing.T) {
	config := main.PrepareConfig()
	require.NotNil(t, config)
	assert.Equal(t, defaultConfigFromFlags, config)
}

func TestWithDockerCompose(t *testing.T) {
	compose := testcontainers.NewLocalDockerCompose([]string{"../../docker-compose.yml"}, "id")
	go func() {
		compose.WithCommand([]string{"up"}).Invoke()
	}()
	defer compose.Down()
	serverAddr := "http://localhost:8080"
	healthCheck(t, serverAddr, 20*time.Second)
	checkApplication(t, serverAddr)
}

func checkApplication(t *testing.T, serverAddr string) {
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
	// request balances to trigger the cache
	user1.requestBalance(t)
	user2.requestBalance(t)
	user3.requestBalance(t)
	// start paying, that should clean the cache for whom it is needed
	user1.payForPizza(t)
	user2.payForCoffee(t)
	checkBalances(t, user1, user2, user3)
}
