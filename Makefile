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
	@CONFIG_PATH=$(PREFIX)/_config/worker.yml go run sandman-scheduler/main.go -db-migrate -start=false

migrate:
	@CONFIG_PATH=$(PREFIX)/_config/worker.yml go run sandman-scheduler/main.go -db-migrate -start=false

run:
	@CONFIG_PATH=$(PREFIX)/_config/server.yml go run dev/main.go