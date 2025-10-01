.PHONY: run, migrate

STEPS ?= 0

ifneq (,$(wildcard ./.env))
    include .env
    export
endif

run:
	go run cmd/spiry/main.go

migrate:
	go run cmd/spiry/migrate.go --steps $(STEPS)