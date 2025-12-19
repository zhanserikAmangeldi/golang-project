# Используем 1.24, так как 1.25 еще может быть недоступна в стабильных репозиториях
FROM golang:1.24-alpine AS builder

# Разрешаем Go скачивать нужную версию тулчейна, если в go.mod указана выше
ENV GOTOOLCHAIN=auto

WORKDIR /app

# Устанавливаем swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Сначала копируем только файлы зависимостей для кэширования слоев
COPY go.mod go.sum ./
RUN go mod download

# Теперь копируем остальной код
COPY . .

# Генерируем документацию
RUN swag init -g cmd/api/main.go

# Сборка
RUN go build -o user-service ./cmd/api/main.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/user-service .
COPY --from=builder /app/internal/migration/migrations ./internal/migration/migrations
COPY --from=builder /app/docs ./docs

EXPOSE 8081 9091

CMD ["./user-service"]