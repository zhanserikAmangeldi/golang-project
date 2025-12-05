FROM golang:1.24-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git


COPY proto/ /proto/


COPY user-service/go.mod user-service/go.sum ./


RUN go mod download

COPY user-service/ ./


RUN go install github.com/swaggo/swag/cmd/swag@latest


RUN swag init -g cmd/api/main.go -o cmd/api/docs


RUN go build -o user-service ./cmd/api

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/user-service .

EXPOSE 8081

CMD ["./user-service"]