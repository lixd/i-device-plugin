IMG = docker.io/i-device-plugin:latest

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/i-device-plugin cmd/main.go

.PHONY:build-image
build-image:
	docker build -t ${IMG} .