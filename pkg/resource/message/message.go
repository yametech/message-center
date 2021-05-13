package message

import (
	"github.com/yametech/devops/pkg/core"
	"github.com/yametech/devops/pkg/store"
	"github.com/yametech/devops/pkg/store/gtm"
)

const Kind core.Kind = "message"

type Status uint8

const (
	Init Status = iota
	Fail
	Success

)

type Spec struct {
	SendUser []string `json:"send_user" bson:"send_user"`
	MsgType  string   `json:"msg_type" bson:"msg_type"`
	Content  string   `json:"content" bson:"content"`
	Status
}

type Message struct {
	core.Metadata `json:"metadata"`
	Spec          Spec `json:"spec"`
}

type ReqDepart struct {
	ErrCode    int   `json:"errcode"`
	DeptIdList []int `json:"sub_dept_id_list"`
}

type ReqUser struct {
	ErrCode  int `json:"errcode"`
	UserList []struct {
		Name   string `json:"name"`
		Userid string `json:"userid"`
	} `json:"userlist"`
}

// Pipeline impl Coder
func (*Message) Decode(op *gtm.Op) (core.IObject, error) {
	message := &Message{}
	if err := core.ObjectToResource(op.Data, message); err != nil {
		return nil, err
	}
	return message, nil
}

func init() {
	store.AddResourceCoder(string(Kind), &Message{})
}
