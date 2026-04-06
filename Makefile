.PHONY: build run test lint clean

build:
	mkdir -p bin
	go build -o bin/main .

run: build
	./bin/main

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -f ./bin/main
