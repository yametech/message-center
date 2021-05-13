package message

import "github.com/yametech/devops/pkg/core"

type UserSpec struct {
	Name string `json:"name"`
	DingID string `json:"ding_id" bson:"ding_id"`
}

type User struct {
	core.Metadata `json:"metadata"`
	Spec          UserSpec `json:"spec"`
}
