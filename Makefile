.PHONY: run, build, docs test

STEPS ?= 0

ifneq (,$(wildcard ./.env))
    include .env
    export
endif

run:
	make build
	./app.out --steps $(STEPS)

build:
	go build -o ./app.out cmd/spiry/main.go

test:
	go test -v ./...

docs:
	swag init -g ../cmd/spiry/main.go -d ./internal -o ./docs