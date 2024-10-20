PREFIX ?= $(shell pwd)


init: ensure-protoc-gen-go ensure-protoc-gen-go-grpc

ensure-protoc-gen-go:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

ensure-protoc-gen-go-grpc:
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

generate:
	@go generate ./...

test:
	@go test ./...

db:
	@cockroach sql --insecure --execute="drop database if exists sandman"
	@cockroach sql --insecure --execute="create database sandman"
	@CONFIG_PATH=$(PREFIX)/_config/worker.yml go run sandman-worker/main.go -db-migrate -start=false

server:
	@CONFIG_PATH=$(PREFIX)/_config/worker.yml go run sandman-srv/main.go

worker:
	@CONFIG_PATH=$(PREFIX)/_config/worker.yml go run sandman-worker/main.go