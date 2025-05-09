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

tools:
	go get -u mvdan.cc/gofumpt@latest
	go get -u github.com/daixiang0/gci@latest

lint: tidy
	gofumpt -w .
	gci write . --skip-generated -s standard -s default
	golangci-lint run ./...

test: up
	go test -race ./... -v -coverpkg=./... -coverprofile=coverage.txt -covermode atomic
	go tool cover -func=coverage.txt | grep 'total'
	which gocover-cobertura || go install github.com/t-yuki/gocover-cobertura@latest
	gocover-cobertura < coverage.txt > coverage.xml

.PHONY: build
build:
	GOOS=linux go build -o bin/loader ./cmd/load-json/main.go
	GOOS=linux go build -o bin/service ./cmd/clothing-service/main.go

deploy:
	scp ./bin/loader adam@84.201.181.17:/home/adam/back/loader