# Expense tracking application backend

To see API please check `openapi.yml`.

## Build

To build an application execute:

```
make app
```

If using make is not an option please do:

```
go build ./cmd/go-spend
```

To build docker image do:

```
make docker
```

Or alternatively:

```
docker build -t go-spend .
```

## Tests

The tests use `Docker` heavily through the [go-testcontainers](https://github.com/testcontainers/testcontainers-go)
library. So in order to execute all of them properly `Docker` is required.
E2E test requires usage of `docker-compose` otherwise it will fail.

First test execution might be a bit slow - images necessary for the test will be downloaded. That includes postgres and
redis. Consequent executions will be faster but still require some time for containers to start.

There are unit tests, integration test, and some simple e2e test. Their execution is not separated.

To execute test do:

```
make test
```

Or:

```
go test ./...
```

If test execution is slow on docker-related tests please check VPN and network.

## Notes

The simplest email check was added - it doesn't support multiple subdomains.
