GO        ?= go
COMPOSE   ?= docker compose
COMPOSE_F := -f local/docker-compose.yaml

.PHONY: build test vet lint tidy run-api up down logs ps smoke

build:
	$(GO) build ./...

test:
	$(GO) test ./... -race -count=1

vet:
	$(GO) vet ./...

lint:
	golangci-lint run

tidy:
	$(GO) mod tidy

run-api:
	ALERTER_LOG_FORMAT=text $(GO) run ./services/api/cmd/api

up:
	$(COMPOSE) $(COMPOSE_F) up --build -d

down:
	$(COMPOSE) $(COMPOSE_F) down

logs:
	$(COMPOSE) $(COMPOSE_F) logs -f

ps:
	$(COMPOSE) $(COMPOSE_F) ps

smoke:
	curl -fsS http://localhost:8080/v1/ping && echo
