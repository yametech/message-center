package utils

import (
	"github.com/yametech/message-center/pkg/resource/message"
)

type GoMessage struct {
	Title    string
	Content  string
	Receiver []string
}

func (gm *GoMessage) Build() *message.Message {
	msg := &message.Message{
		Spec: message.Spec{
			SendUser: gm.Receiver,
			Content:  gm.Content,
			Title:    gm.Title,
		},
	}
	msg.GenerateVersion()
	return msg
}
