.PHONY: build \
		run \
		test test-with-coverage \
		lint \
		clean \
		docker-build docker-run

IMAGE_NAME ?= go-project-278
CONTAINER_NAME ?= go-project-278

build:
	mkdir -p bin
	go build -o bin/main .

run: build
	./bin/main

test:
	go test ./...

test-with-coverage:
	go test -coverprofile=coverage.out ./...

lint:
	golangci-lint run

clean:
	rm -f ./bin/main

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run --rm \
		--name $(CONTAINER_NAME) \
		-p 8080:8080 \
		-e PORT=8080 \
		-e SENTRY_DSN="$(SENTRY_DSN)" \
		$(IMAGE_NAME)
		