package message

import "github.com/yametech/devops/pkg/core"

type Spec struct {
	SendUser []string `json:"send_user" bson:"send_user"`
	MsgType  string   `json:"msg_type" bson:"msg_type"`
	Content  string   `json:"content" bson:"content"`
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
	ErrCode  int    `json:"errcode"`
	UserList []struct {
		Name   string `json:"name"`
		Userid string `json:"userid"`
	} `json:"userlist"`
}
