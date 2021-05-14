package utils

import (
	"github.com/yametech/devops/pkg/store/mongo"
	"github.com/yametech/message-center/pkg/common"
	"testing"
)

func TestGenerate(t *testing.T) {
	store, err, _ := mongo.NewMongo("mongodb://10.200.10.46:27017/admin")
	if err != nil {
		t.Fatal(err)
	}
	message := GoMessage{
		Title:    "DXP IS SB",
		Receiver: []string{"李佳明"},
		Content:  "#### nothing",
	}
	obj := message.Build()
	_, _, err = store.Apply(common.DefaultNamespace, common.MessageCenter, obj.UUID, obj, false)
	if err != nil {
		t.Fatal(err)
	}

}
