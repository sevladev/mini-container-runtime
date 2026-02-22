.PHONY: shell build install test lint clean

shell:
	docker compose run --rm dev bash

build:
	docker compose run --rm dev go build -o bin/minic ./cmd/minic

install:
	docker compose run --rm dev sh -c "go build -o /usr/local/bin/minic ./cmd/minic"

test:
	docker compose run --rm dev go test ./... -v

lint:
	docker compose run --rm dev go vet ./...

clean:
	rm -rf bin/
