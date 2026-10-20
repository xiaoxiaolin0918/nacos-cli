VERSION ?= 0.3.0
PKG := github.com/xiaoxiaolin0918/nacos-cli
LDFLAGS := -X $(PKG)/internal/version.Version=$(VERSION)

.PHONY: build test tidy fmt version

version:
	@echo $(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o nacos.exe .

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .
