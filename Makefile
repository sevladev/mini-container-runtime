.PHONY: shell build test lint clean

shell:
	docker compose run --rm dev bash

build:
	docker compose run --rm dev go build -o bin/minic ./cmd/minic

test:
	docker compose run --rm dev go test ./... -v

lint:
	docker compose run --rm dev go vet ./...

clean:
	rm -rf bin/
