#!/usr/bin/env bash

docker rm postgres || true

docker run --name postgres \
-e POSTGRES_USER=user \
-e POSTGRES_PASSWORD=password \
-e POSTGRES_DB=expenses \
-p 5432:5432 \
-v "$(pwd)"/db:/docker-entrypoint-initdb.d \
postgres:13.1
