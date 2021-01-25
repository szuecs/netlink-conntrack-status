VERSION     ?= $(shell git describe --tags --always --dirty)
COMMIT_HASH ?= $(shell git rev-parse --short HEAD)

default: build

build:
	GO111MODULE=$(GO111) go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT_HASH)" -o netlink-conntrack-status .
