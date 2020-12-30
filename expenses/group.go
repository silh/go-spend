package expenses

import (
	"bytes"
	"encoding/json"
	"go-spend/util"
)

// Group as it is present in DB. Every User can be only on one group.
type Group struct {
	ID   uint
	Name util.NonEmptyString
}

type GroupResponse struct {
	ID    uint                `json:"id"`
	Name  util.NonEmptyString `json:"name"`
	Users []UserResponse      `json:"users"`
}

// A request to create a Group
type CreateGroupRequest struct {
	Name      util.NonEmptyString `json:"name"`
	CreatorID uint                `json:"creatorId"`
}

// UnmarshalJSON transforms the request JSON data and validates it.
func (c *CreateGroupRequest) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	type createGroupRequest struct {
		Name      string `json:"name"`
		CreatorID uint   `json:"creatorId"`
	}
	var req createGroupRequest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var err error
	if err = decoder.Decode(&req); err != nil {
		return err
	}
	c.Name, err = util.NewNonEmptyString(req.Name)
	if err != nil {
		return err
	}
	c.CreatorID = req.CreatorID
	return nil
}
