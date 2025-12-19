## Chat Application: Microservices Architecture Tutorial
Welcome to the Real-Time Chat Application repository. This project is a production-ready demonstration of a scalable microservices ecosystem built using Go (Golang).

This tutorial will guide you through the system design, the technology stack, and how to deploy the entire environment using Docker.

### System Architecture
The project follows a decoupled microservices approach. Instead of a single "monolith," the responsibilities are split between specialized services.

<img width="361" height="551" alt="Диаграмма без названия drawio (1)" src="https://github.com/user-attachments/assets/606d04bf-9f56-4880-a09d-6a696ffa2940" />

1. API Gateway

Role: Acts as a reverse proxy.

Function: Routes HTTP and WebSocket requests to the correct service. It simplifies the client-side logic by providing one URL for all operations.

2. Server Layer (Microservices)

User-Service: Manages user registration, profile updates, and authentication.

Chat-Service: Manages group creation, message persistence, and real-time broadcasting.

Inter-service Communication: Services talk to each other using gRPC for high-speed data exchange.

3. Data Layer

PostgreSQL (PSQL): Reliable relational storage for users and message history.

Redis: High-speed in-memory data store used for session caching and the Pub/Sub mechanism for real-time messaging.

MinIO: S3-compatible object storage for handling file uploads (avatars, images, documents).

🛠 Prerequisites
Before starting, ensure you have the following installed:

Docker & Docker Compose

Go (1.21 or higher, if running locally)

Swag CLI (for API documentation)

🚀 Getting Started (Step-by-Step)
1. Environment Configuration

Create a .env file in the root directory. This file stores your credentials. Based on the docker-compose.yml, you need to define:

Фрагмент кода
# Database Config
DB_USER=postgres
DB_PASSWORD=your_password
USER_DB_NAME=user_db
CHAT_DB_NAME=chat_db

# Ports
USER_HTTP_PORT=8081
CHAT_HTTP_PORT=8082
GATEWAY_PORT=8080

# MinIO & Redis
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=password
REDIS_PORT=6379

2. How to run

1. Git clone our repo and go in root folder:
```
git clone https://github.com/zhanserikAmangeldi/golang-project.git
cd golang-project
```

2. Run the docker compose:
```
docker-compose up --build
```

PS: you can find `.env` file where you can anything

3. Generate API Documentation

If you modify the handlers, you must refresh the Swagger UI:

Bash
# Inside chat-service directory
swag init -g cmd/main.go

# Inside user-service directory
swag init -g cmd/main.go
💬 Real-Time Logic: How it Works
The system integrates HTTP and WebSockets to ensure messages are both permanent and instant.

Request: User A sends a message via POST /api/v1/messages/send.

Storage: The Chat Service saves the message to PostgreSQL.

Broadcasting: The service publishes the message to a Redis channel.

Delivery: All Chat Service instances listening to that Redis channel receive the message and push it to User B via an open WebSocket connection.

```
golang-project/
├── .env                    # Глобальные переменные окружения
├── docker-compose.yml      # Конфигурация всей инфраструктуры
├── api-gateway/            # Прокси-сервер
│   ├── cmd/
│   │   └── main.go         # Точка входа Gateway
│   └── Dockerfile
├── user-service/           # Микросервис пользователей
│   ├── cmd/
│   │   └── main.go         # Главный файл (Swag init -g cmd/main.go)
│   ├── docs/               # Автогенерация Swagger UI
│   ├── internal/           # Бизнес-логика (handlers, services, repositories)
│   ├── proto/              # Protobuf файлы для gRPC
│   └── Dockerfile
├── chat-service/           # Микросервис чатов
│   ├── cmd/
│   │   └── main.go         # Главный файл (Swag init -g cmd/main.go)
│   ├── docs/               # Автогенерация Swagger UI
│   ├── internal/           # Логика чатов, Redis и MinIO адаптеры
│   ├── proto/
│   └── Dockerfile 
```

The User Service Swagger is usually at http://localhost:8082/swagger/index.html.

Database Connection Issues: Check if the containers are "Healthy" using docker-compose ps.

Missing Methods in Swagger: Ensure you have added the @Router annotations in your handler files and ran swag init.

