module github.com/yametech/message-center

go 1.16

require (
	github.com/levigross/grequests v0.0.0-20190908174114-253788527a1a
	github.com/yametech/devops v0.0.0-20210512080738-08e7cb3f01e1
)

replace google.golang.org/grpc v1.29.1 => google.golang.org/grpc v1.26.0
replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

