# Real-Time Chat Application Technology Stack Report

!!!Временно!!!

## Executive Summary

This report focuses on building a real-time chat application in Go using a practical, learning-oriented approach with Docker Compose and simple microservices. The core technologies are Gorilla WebSocket, gRPC, RabbitMQ, Redis, and PostgreSQL.

## Core Technologies

### WebSocket Communication
**Gorilla WebSocket**
- Real-time bidirectional communication with clients
- Excellent performance and stability
- Rich feature set for connection management

### Service Communication
**gRPC**
- High-performance RPC for inter-service communication
- Protocol Buffers for efficient serialization
- Built-in streaming support for real-time features
- Strong Go support

### Message Queuing
**RabbitMQ**
- Reliable message delivery with exchanges and queues
- Excellent routing capabilities
- Strong durability and delivery guarantees
- Great Go client library

### Database & Caching
**PostgreSQL**
- Message persistence and user data
- Excellent JSON support and performance
- Strong consistency

**Redis**
- User sessions and authentication tokens
- User presence tracking
- Rate limiting

## Architecture

### Docker Compose Setup

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: chatapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
  
  rabbitmq:
    image: rabbitmq:3-management
    environment:
      RABBITMQ_DEFAULT_USER: admin
      RABBITMQ_DEFAULT_PASS: admin123
    ports:
      - "5672:5672"
      - "15672:15672" # Management UI
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq

  chat-api:
    build: ./chat-service
    ports:
      - "8080:8080"      # HTTP/WebSocket
      - "9090:9090"      # gRPC client port
    environment:
      - GRPC_USER_SERVICE=user-api:9091
    depends_on:
      - postgres
      - redis
      - rabbitmq
  
  user-api:
    build: ./user-service
    ports:
      - "8081:8081"      # HTTP
      - "9091:9091"      # gRPC server port
    depends_on:
      - postgres
      - redis

volumes:
  postgres_data:
  rabbitmq_data:
```

### Service Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Frontend/Web   │    │  Chat Service   │    │  User Service   │
│     Client      │────│   (WebSocket)   │────│   (gRPC Server) │
│   (WebSocket)   │    │  (gRPC Client)  │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                    ┌─────────────────┐    ┌─────────────────┐
                    │   PostgreSQL    │    │     Redis       │
                    │   (Messages)    │    │  (Sessions)     │
                    └─────────────────┘    └─────────────────┘
                                │
                    ┌─────────────────┐
                    │   RabbitMQ      │
                    │ (Message Queue) │
                    │  (Exchanges &   │
                    │    Queues)      │
                    └─────────────────┘
```

### RabbitMQ Message Flow

```
Chat Service A ──→ RabbitMQ Exchange ──→ Queue ──→ Chat Service B
                        │
                        ├──→ Message History Service
                        │
                        └──→ Notification Service (optional)
```

## Implementation Phases

### Phase 1: Core Setup (1-2 weeks)
- Docker Compose with PostgreSQL, Redis, RabbitMQ
- Basic Go project structure with two services
- **Gorilla WebSocket** connection in chat service
- **gRPC** setup between chat and user services (basic .proto files)
- Simple user authentication with JWT
- Basic message sending through WebSocket

### Phase 2: Add Queue & gRPC Communication (2-3 weeks)
- **RabbitMQ** integration with exchanges and queues
- **gRPC** service calls for user validation and data
- Message persistence in PostgreSQL through message queues
- Redis for user sessions and presence tracking
- WebSocket to RabbitMQ to WebSocket message flow

### Phase 3: Advanced Messaging (2-3 weeks)
- RabbitMQ routing for different message types
- Message delivery confirmation through queues
- Chat rooms/channels with dedicated queues
- File upload with message queue processing
- gRPC streaming for real-time user presence updates

**Total Timeline: 5-8 weeks of learning**

## Essential Libraries

```go
// WebSocket
"github.com/gorilla/websocket" 

// gRPC
"google.golang.org/grpc"
"google.golang.org/protobuf/proto"
"google.golang.org/grpc/reflection" // For debugging

// HTTP Framework
"github.com/gin-gonic/gin"

// Database
"github.com/jackc/pgx/v5" // PostgreSQL driver
"github.com/redis/go-redis/v9" // Redis client

// Message Queue
"github.com/rabbitmq/amqp091-go" // RabbitMQ client

// Authentication
"github.com/golang-jwt/jwt/v5" // JWT tokens

// Validation & Config
"github.com/go-playground/validator/v10"
"github.com/joho/godotenv"

// Logging
"log/slog" // Go 1.21+

// Database Migrations
"github.com/golang-migrate/migrate/v4"
```

## Development Tools

- **protoc**: Protocol buffer compiler for gRPC
- **protoc-gen-go**: Go code generator for protobuf
- **protoc-gen-go-grpc**: Go gRPC code generator
- **Evans**: gRPC client for testing: `go install github.com/ktr0731/evans@latest`
- **Air**: Hot reload: `go install github.com/cosmtrek/air@latest`
- **Docker & Docker Compose**: For running services locally

## What You'll Learn

- **WebSocket management**: Real-time bidirectional communication with Gorilla WebSocket
- **gRPC communication**: High-performance RPC between microservices
- **Message queuing**: RabbitMQ exchanges, queues, and routing for reliable message delivery
- **Caching strategies**: Redis for sessions and quick data access  
- **Database operations**: PostgreSQL for persistent data
- **Protocol Buffers**: Efficient serialization for service communication
- **Docker containerization**: All services running in containers
- **Authentication**: JWT-based user sessions across services

## Learning Path

### Week 1-2: Docker & Basic Setup
1. Set up Docker Compose with PostgreSQL, Redis, RabbitMQ
2. Create basic Go project structure for 2 services
3. Get **Gorilla WebSocket** working with simple message echo
4. Create basic **gRPC** .proto files and generate Go code
5. Connect to PostgreSQL and create basic tables

### Week 3-4: gRPC & Message Queuing
1. Implement **gRPC** service calls between chat and user services
2. Set up **RabbitMQ** exchanges and queues for message routing
3. Integrate Redis for user sessions
4. Basic JWT authentication validated through gRPC calls

### Week 5-6: Advanced Messaging
1. WebSocket ↔ RabbitMQ ↔ WebSocket message flow
2. **gRPC streaming** for real-time user presence
3. Message persistence through queue consumers
4. Chat rooms with dedicated RabbitMQ routing

### Week 7-8: Polish & Advanced Features
1. Message delivery confirmation via RabbitMQ
2. File upload processing through message queues
3. **gRPC reflection** for debugging services
4. Error handling and connection recovery for both WebSocket and gRPC

## Security Considerations

### Authentication & Authorization
- JWT token validation across services
- Rate limiting with Redis
- Input validation and sanitization
- Secure WebSocket connections (WSS)

### Data Protection
- Parameterized queries for SQL injection prevention
- Message content validation
- Token expiration and refresh strategies

## Conclusion

This focused approach centers on learning five core technologies through hands-on development: Gorilla WebSocket for real-time communication, gRPC for efficient service communication, RabbitMQ for reliable message queuing, Redis for caching, and PostgreSQL for persistence.

Using Docker Compose for local development and simple microservices architecture provides practical experience with modern Go backend patterns while maintaining manageable complexity perfect for learning.