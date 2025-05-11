#!/usr/bin/env bash


go build -o dstream-ingester-time ./cmd/dstream-ingester-time
go run ./cmd/run-time-plugin/main.go
