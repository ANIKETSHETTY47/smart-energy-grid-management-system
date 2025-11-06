SHELL := /bin/bash
APP ?= api
.PHONY: dev build run test ingestor simulate lint fmt

dev:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

run: build
	./bin/api

ingestor:
	go run ./cmd/ingestor

simulate:
	go run ./cmd/simulator

test:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run || true

fmt:
	gofmt -w .
