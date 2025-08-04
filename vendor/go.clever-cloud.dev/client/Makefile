.PHONY: test
test:
	go test -v -covermode=atomic -coverprofile=coverage.out ./...

.PHONY: deps
deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
	go mod vendor

.PHONY: ci
ci: deps lint test

.PHONY: lint
lint:
	go fmt ./...
	golangci-lint run
