#!/usr/bin/env bash


go build -o out/dstream-ingester-time ./cmd/dstream-ingester-time
go run ./cmd/run-time-plugin/main.go
