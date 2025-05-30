.PHONY: generate test

generate:
	go run github.com/vektra/mockery/v3@v3.3.1

test:
	go test --race  ./...