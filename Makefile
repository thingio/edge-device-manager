APP=edge-device-manager
PKG=github.com/thingio/${APP}

VERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)
GIT_SHA=$(shell git rev-parse HEAD)
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

GOMODULE=GO111MODULE=on GOPROXY=https://goproxy.cn,https://goproxy.io/,https://mirrors.aliyun.com/goproxy,direct
GO=CGO_ENABLED=0 $(GOMODULE) go
CGO=CGO_ENABLED=1 $(GOMODULE) go
GOFLAGS=-ldflags "-X $(PKG)/version.Version=$(VERSION)"

.PHONY: update build
update:
	$(GO) mod download all
build:
	$(GO) build $(GOFLAGS) -o $(APP) ./main.go

.PHONY: dockerfile
dockerfile:
	docker build -f Dockerfile . \
	-t $(APP):$(GIT_SHA) \
	-t $(APP):$(GIT_BRANCH) \
	-t $(APP):$(VERSION)