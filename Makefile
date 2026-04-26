PREFIX ?= $(shell pwd)

# `make db` only needs a single live mikoshi node for the DROP/CREATE DDL
# (cluster-wide DDL propagates from any node). Override if this one is
# down: `make db MIKOSHI_HOST=127.0.0.1:26258`. The running services read
# the full `dbHosts` list from _config/config.yml and don't use this.
MIKOSHI_HOST ?= 127.0.0.1:26257

init: ensure-protoc-gen-go ensure-protoc-gen-go-grpc

ensure-protoc-gen-go:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

ensure-protoc-gen-go-grpc:
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

generate:
	@go generate ./...

test:
	@go test ./...

load-test:
	@go run scripts/load_test/main.go

db:
	@mikoshi sql --insecure --host=$(MIKOSHI_HOST) --execute="drop database if exists sandman"
	@mikoshi sql --insecure --host=$(MIKOSHI_HOST) --execute="create database sandman"
	@CONFIG_PATH=$(PREFIX)/_config/config.yml go run sandman-worker/main.go -db-migrate -start=false

migrate:
	@CONFIG_PATH=$(PREFIX)/_config/config.yml go run sandman-worker/main.go -db-migrate -start=false

run:
	@CONFIG_PATH=$(PREFIX)/_config/config.yml go run scripts/dev/main.go

run-worker-00:
	@CONFIG_PATH=$(PREFIX)/_config/config.yml EXPVAR_LISTEN_ADDR=:8082 HOSTNAME=worker-00 go run sandman-worker/main.go

run-worker-01:
	@CONFIG_PATH=$(PREFIX)/_config/config.yml EXPVAR_LISTEN_ADDR=:8083 HOSTNAME=worker-01 go run sandman-worker/main.go
