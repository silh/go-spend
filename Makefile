all: fmt
	@go build ./cmd/go-spend

fmt:
	@go fmt ./...
