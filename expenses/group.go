package expenses

import "go-spend/util"

// Group as it is present in DB. Every User can be only on one group.
type Group struct {
	ID   uint
	Name util.NonEmptyString
}

// A request to create a group
type CreateGroupRequest struct {
	Name      util.NonEmptyString `json:"name"`
	CreatorID uint                `json:"creatorId"`
}
