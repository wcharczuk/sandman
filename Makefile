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
	@CONFIG_PATH=$(PREFIX)/_config/config.yml go run sandman-worker/main.go -db-migrate -start=false

migrate:
	@CONFIG_PATH=$(PREFIX)/_config/config.yml go run sandman-worker/main.go -db-migrate -start=false

run:
	@CONFIG_PATH=$(PREFIX)/_config/config.yml go run scripts/dev/main.go

run-worker-00:
	@CONFIG_PATH=$(PREFIX)/_config/config.yml EXPVAR_BIND_ADDR=:8082 HOSTNAME=worker-00 go run sandman-worker/main.go

run-worker-01:
	@CONFIG_PATH=$(PREFIX)/_config/config.yml EXPVAR_BIND_ADDR=:8083 HOSTNAME=worker-00 go run sandman-worker/main.go
