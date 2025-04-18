run:
	@echo 'Running the project ...'
	go build -o bin/main ./cmd/clothing-service/main.go
	./bin/main

up: 
	docker compose up -d

down:
	docker compose down --remove-orphans

tidy:
	go mod tidy

lint: tidy
	gofumpt -w .
	gci write . --skip-generated -s standard -s default
	golangci-lint run ./...
