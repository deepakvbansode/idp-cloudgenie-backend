
PACKAGE = ${shell pwd | rev | cut -f1 -d'/' - | rev}
DATE  ?= ${shell  date +%Y-%m-%d_%I:%M:%S%p}
GITHASH = ${shell git rev-parse HEAD}
DOCKER_BUILD_CONTEXT =.
DOCKER_FILE_PATH =Dockerfile
DOCKER_BUILD_ARGS =--build-arg GOARCH=amd64 --build-arg GOOS=linux
DOCKER_BUILD_ARGS_MAC =--build-arg GOARCH=arm64 --build-arg GOOS=linux
.DEFAULT_GOAL := all

.PHONY: all
all: fmt build

.PHONY: fmt
fmt: 
	go fmt ./...


build:
	go build -tags release -ldflags '-X main.GitComHash=$(GITHASH) -X main.BuildStamp=$(DATE)' -o bin/application cmd/server/main.go

docker:
	@docker build $(DOCKER_BUILD_ARGS) -t $(PACKAGE):$(GITHASH) $(DOCKER_BUILD_CONTEXT) -f $(DOCKER_FILE_PATH)
	@docker tag $(PACKAGE):$(GITHASH) $(PACKAGE):latest
	@docker tag $(PACKAGE):$(GITHASH) $(CI_REGISTRY_IMAGE)

docker-mac:
	@docker build $(DOCKER_BUILD_ARGS_MAC) -t $(PACKAGE):$(GITHASH)-arm64 $(DOCKER_BUILD_CONTEXT) -f $(DOCKER_FILE_PATH)
	@docker tag $(PACKAGE):$(GITHASH)-arm64  deepakbansode/$(PACKAGE):latest-arm64 
