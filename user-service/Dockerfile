FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o chat-service ./cmd/api/main.go

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/chat-service .


EXPOSE 8081 9091

CMD ["./chat-service"]
