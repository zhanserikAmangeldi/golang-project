FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install git
RUN apk add --no-cache git


COPY proto/ /proto/


# -------------------------------------
COPY chat-service/go.mod chat-service/go.sum ./


RUN go mod download


COPY chat-service/ ./


RUN go build -o chat-service ./cmd


FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/chat-service .

EXPOSE 8080

CMD ["./chat-service"]