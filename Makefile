BINARY_NAME=github.com/ThembinkosiThemba/zen
SOURCES=$(wildcard *.go)

build:
	go build -o $(BINARY_NAME)

tidy:
	go mod tidy

run:
	go run docs/usage/main.go
	
swagger:
	cd cmd && swag init

test:
	go test -list .

clean:
	rm -f $(BINARY_NAME)

# Define phony targets (these don't represent files)
.PHONY: all build run test clean
