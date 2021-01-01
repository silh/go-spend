app: fmt
	@go build ./cmd/go-spend

test: fmt
	@go test ./...

fmt:
	@go fmt ./...

docker:
	@docker build -t go-spend .
