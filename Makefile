GOLANGCI_VERSION ?= v1.51.0
DOCKER_COMPOSE ?= docker compose

.PHONY: lint
lint:
	docker run --rm -it --network none -v $(CURDIR):/app -w /app golangci/golangci-lint:$(GOLANGCI_VERSION) golangci-lint run

.PHONY: format
format:
	gci write graph internal pkg server.go
	gofumpt -w -extra .

.PHONY: vendor
vendor:
	go get -v ./...
	go mod tidy
	go mod vendor

.PHONY: test
test:
	go test -count=1 -v ./...
