init: ensure-protoc-gen-go ensure-protoc-gen-go-grpc

ensure-protoc-gen-go:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

ensure-protoc-gen-go-grpc:
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

generate:
	@go generate ./...

test:
	@go test ./...
