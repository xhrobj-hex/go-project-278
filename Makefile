.PHONY: build \
		clean \
		postgres-up postgres-start postgres-stop postgres-rm postgres-connect \
		run \
		test test-with-coverage \
		lint \
		docker-build docker-run \
		migrate-up migrate-down migrate-status

POSTGRES_USER=shorty
POSTGRES_PASSWORD=secret
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=shortener
POSTGRES_DSN=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

IMAGE_NAME ?= go-project-278
CONTAINER_NAME ?= go-project-278

build:
	mkdir -p bin
	go build -o bin/main .

clean:
	rm -f ./bin/main

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
	DATABASE_URL=$(POSTGRES_DSN) ./bin/main

test:
	go test ./...

test-with-coverage:
	go test -coverprofile=coverage.out ./...

lint:
	golangci-lint run

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run --rm \
		--name $(CONTAINER_NAME) \
		-p 8080:8080 \
		-e PORT=8080 \
		-e SENTRY_DSN="$(SENTRY_DSN)" \
		$(IMAGE_NAME)

migrate-up:
	goose -dir migrations postgres "$(POSTGRES_DSN)" up

migrate-down:
	goose -dir migrations postgres "$(POSTGRES_DSN)" down

migrate-status:
	goose -dir migrations postgres "$(POSTGRES_DSN)" status
