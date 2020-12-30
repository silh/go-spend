package expenses_test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"go-spend/util"
	"strings"
	"testing"
)

func TestCreateGroupRequestUnmarshalJSON(t *testing.T) {
	groupJSON := `{"name":"name", "creatorId": 1}`
	var result expenses.CreateGroupRequest
	err := json.NewDecoder(strings.NewReader(groupJSON)).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, util.NonEmptyString("name"), result.Name)
	assert.Equal(t, uint(1), result.CreatorID)
}

func TestCreateGroupRequestUnmarshalJSONNull(t *testing.T) {
	groupJSON := "null"
	var result expenses.CreateGroupRequest
	err := json.NewDecoder(strings.NewReader(groupJSON)).Decode(&result)
	require.NoError(t, err)
	assert.Zero(t, result)
}

func TestCreateGroupRequestUnmarshalJSONErrors(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "unexpected fields",
			json: `{"name":"name", "creatorId": 1, "zz":"aaaa"}`,
		},
		{
			name: "no name field",
			json: `{"name":"name", "creatorId": 1, "zz":"aaaa"}`,
		},
		{
			name: "empty name",
			json: `{"name":"", "creatorId": 1}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var req expenses.CreateGroupRequest
			err := json.Unmarshal([]byte(test.json), &req)
			require.Error(t, err)
		})
	}
}
