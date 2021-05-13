all: build-appservice-controller


build-appservice-controller:
	docker build -t harbor.ym/devops/message-center:v0.0.1 -f Dockerfile .
	docker push harbor.ym/devops/message-center:v0.0.1