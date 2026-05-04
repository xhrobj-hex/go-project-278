.PHONY: build \
		postgres-up postgres-start postgres-stop postgres-rm postgres-connect \
		run \
		migrate-status \
		lint \
		test test-coverage \
		docker-build docker-run \
		clean \

POSTGRES_USER=shorty
POSTGRES_PASSWORD=secret
POSTGRES_PORT=5432
POSTGRES_DB=shortener

LOCAL_POSTGRES_HOST=localhost
LOCAL_POSTGRES_DSN=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(LOCAL_POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

DOCKER_POSTGRES_HOST ?= host.docker.internal
DOCKER_POSTGRES_DSN=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(DOCKER_POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

IMAGE_NAME ?= go-project-278
CONTAINER_NAME ?= go-project-278
SENTRY_DSN ?=

build:
	mkdir -p bin
	go build -o bin/main .

postgres-up:
	docker run --name shortener-postgres \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):5432 \
		-d postgres:16

postgres-start:
	docker start shortener-postgres

postgres-stop:
	docker stop shortener-postgres

postgres-rm:
	docker rm -f shortener-postgres

postgres-connect:
	docker exec -it shortener-postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

run: build
	DATABASE_URL=$(LOCAL_POSTGRES_DSN) ./bin/main

migrate-status:
	goose -dir migrations postgres "$(LOCAL_POSTGRES_DSN)" status

lint:
	golangci-lint run

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run --rm \
		--name $(CONTAINER_NAME) \
		-p 8080:8080 \
		-e DATABASE_URL=$(DOCKER_POSTGRES_DSN) \
		-e PORT=8080 \
		-e SENTRY_DSN="$(SENTRY_DSN)" \
		$(IMAGE_NAME)

clean:
	rm -f ./bin/main
