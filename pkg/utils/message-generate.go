package utils

import (
	"fmt"
	"github.com/yametech/message-center/pkg/resource/message"
)

type GoMessage struct {
	Title    string
	Content  string
	Receiver []string
}

func (gm *GoMessage) Generate() string {
	data := fmt.Sprintf("#### %s\n\n", gm.Title)
	data = fmt.Sprintf("%s%s", data, gm.Content)
	return data
}

func (gm *GoMessage) Build() *message.Message {
	msg := &message.Message{
		Spec: message.Spec{
			SendUser: gm.Receiver,
			Content:  gm.Generate(),
		},
	}
	msg.GenerateVersion()
	return msg
}
