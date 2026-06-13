.PHONY: test-unit test-integration bench run migrate-up build docker-up docker-migrate-up docker-down

test-unit:
	go test ./... -short -count=1

test-integration:
	go test ./... -tags=integration -count=1

bench:
	go test -bench=. -benchmem ./...

run:
	go run . --config config/config.example.yaml server

migrate-up:
	go run . --config config/config.example.yaml migrate up

build:
	go build -o bin/monaire-todo .

docker-migrate-up:
	docker compose run --rm api migrate up

docker-up:
	docker compose up --build

docker-down:
	docker compose down
