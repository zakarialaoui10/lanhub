.PHONY: fmt test tidy example

fmt:
	gofmt -w .

test:
	go test ./...

tidy:
	go mod tidy

example:
	go run ./examples/basic
