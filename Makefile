.PHONY: build run lint clean

build:
	mkdir -p bin
	go build -o bin/main .

run: build
	./bin/main

lint:
	golangci-lint run

clean:
	rm -f ./bin/main
