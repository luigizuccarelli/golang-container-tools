.PHONY: all test build clean

all: clean test build

build: 
	mkdir -p build
	go build -o build ./...

test:
	mkdir -p tests/results
	go test -v -coverprofile=tests/results/cover.out ./...

cover:
	go tool cover -html=tests/results/cover.out -o tests/results/cover.html

clean:
	rm -rf build/*
	go clean ./...

container:
	podman build -t  quay.io/luzuccar/golang-container-tools:v0.0.1 .

push:
	podman push quay.io/luzuccar/golang-container-tools:v0.0.1
