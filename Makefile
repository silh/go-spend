all: fmt
	@go build ./cmd/go-spend

test:
	@go test ./...

fmt:
	@go fmt ./...
