.PHONY: test race bench vet fmt run

test:
	go test ./...

race:
	go test -race ./...

bench:
	go test -run xxx -bench . -benchmem .

vet:
	go vet ./...

fmt:
	gofmt -w .

run:
	go run ./examples/basic
