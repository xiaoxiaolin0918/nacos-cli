.PHONY: build test tidy
build:
	go build -o nacos.exe .
test:
	go test ./...
tidy:
	go mod tidy
