.PHONY: build run test test-with-coverage lint clean

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
