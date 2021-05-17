# message-center


#### useage
````go
msgUtils "github.com/yametech/message-center/pkg/utils"

msg := msgUtils.GoMessage{
    Title:    "enter your title",
    Receiver: []string{"receiver_1","receiver_2"},
    Content:  "markdown content",
}
sendObj := msg.build()
mongo.apply(sendObj)

````


