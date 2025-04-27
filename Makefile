run:
	@echo 'Running the project ...'
	go build -o bin/main ./cmd/clothing-service/main.go
	./bin/main

up: 
	docker compose -f deployment/local/docker-compose.yml up -d

down:
	docker compose -f deployment/local/docker-compose.yml down --remove-orphans

tidy:
	go mod tidy

lint: tidy
	gofumpt -w .
	gci write . --skip-generated -s standard -s default
	golangci-lint run ./...
