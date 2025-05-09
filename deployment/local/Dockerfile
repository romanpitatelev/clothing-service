FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY ../..go.mod ../../go.sum ./
RUN go mod download

COPY ../.. .
RUN go build -o clothing-service ./cmd/clothing-service/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/variant-service .

EXPOSE 8081
CMD ["./clothing-service"]