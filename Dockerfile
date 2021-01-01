FROM golang:1.15.5-alpine AS build
WORKDIR /src
COPY . .
RUN go build ./cmd/go-spend
FROM alpine:3.12.3 AS bin
COPY --from=build /src /
CMD /go-spend
