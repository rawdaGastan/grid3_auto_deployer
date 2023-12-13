PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)

all: lint test

build:
	@echo "Running $@"
	@go build -ldflags="-X 'github.com/codescalers/cloud4students/cmd.Commit=$(shell git rev-parse HEAD)'" -o bin/cloud4students main.go

run: build
	@echo "Running $@"
	bin/cloud4students 

test: 
	@echo "Running Tests"
	go test -v ./...

coverage: clean 
	mkdir coverage
	go test -v -vet=off ./... -coverprofile=coverage/coverage.out
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@${GOPATH}/bin/gopherbadger -png=false -md="README.md"
	rm coverage.out

clean:
	rm ./coverage -rf
	rm ./bin -rf
	
lint:
	@echo "Installing golangci-lint" && go get github.com/golangci/golangci-lint/cmd/golangci-lint && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	go mod tidy
	@echo "Running $@"
	${GOPATH}/bin/golangci-lint run -c .golangci.yml --timeout 10m
