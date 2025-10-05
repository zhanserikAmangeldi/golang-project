# Structure

realtime-chat-app/
│
├── docker-compose.yml
├── .env.example
├── Makefile
├── README.md
│
├── proto/                          # Shared protobuf definitions
│   ├── user/
│   │   └── user.proto
│   ├── chat/
│   │   └── chat.proto
│   └── common/
│       └── common.proto
│
├── chat-service/                   # Chat/WebSocket Service
│   ├── cmd/
│   │   └── main.go                # Service entry point
│   │
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go          # Service configuration
│   │   │
│   │   ├── handler/
│   │   │   ├── websocket.go       # WebSocket connection handler
│   │   │   ├── chat.go            # Chat HTTP handlers
│   │   │   └── health.go          # Health check endpoint
│   │   │
│   │   ├── websocket/
│   │   │   ├── client.go          # WebSocket client management
│   │   │   ├── hub.go             # Connection hub/manager
│   │   │   └── message.go         # Message types & handling
│   │   │
│   │   ├── service/
│   │   │   ├── chat.go            # Business logic for chat
│   │   │   ├── message.go         # Message processing
│   │   │   └── room.go            # Chat room management
│   │   │
│   │   ├── repository/
│   │   │   ├── message.go         # Message persistence
│   │   │   ├── room.go            # Room data access
│   │   │   └── redis.go           # Redis operations
│   │   │
│   │   ├── grpc/
│   │   │   └── client.go          # gRPC client for user service
│   │   │
│   │   ├── queue/
│   │   │   ├── publisher.go       # RabbitMQ publisher
│   │   │   ├── consumer.go        # RabbitMQ consumer
│   │   │   └── handler.go         # Queue message handlers
│   │   │
│   │   ├── middleware/
│   │   │   ├── auth.go            # JWT validation middleware
│   │   │   ├── ratelimit.go       # Rate limiting
│   │   │   └── cors.go            # CORS configuration
│   │   │
│   │   └── models/
│   │       ├── message.go         # Message domain model
│   │       ├── room.go            # Room domain model
│   │       └── user.go            # User domain model
│   │
│   ├── pkg/                       # Public packages
│   │   ├── logger/
│   │   │   └── logger.go
│   │   └── validator/
│   │       └── validator.go
│   │
│   ├── migrations/                # Database migrations
│   │   ├── 000001_create_messages_table.up.sql
│   │   ├── 000001_create_messages_table.down.sql
│   │   ├── 000002_create_rooms_table.up.sql
│   │   └── 000002_create_rooms_table.down.sql
│   │
│   ├── pb/                        # Generated protobuf code
│   │   ├── user/
│   │   └── chat/
│   │
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── .air.toml                  # Hot reload config
│   └── .env
│
├── user-service/                  # User/Auth Service
│   ├── cmd/
│   │   └── main.go                # Service entry point
│   │
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go          # Service configuration
│   │   │
│   │   ├── handler/
│   │   │   ├── user.go            # User HTTP handlers
│   │   │   ├── auth.go            # Authentication handlers
│   │   │   └── health.go          # Health check endpoint
│   │   │
│   │   ├── service/
│   │   │   ├── user.go            # User business logic
│   │   │   ├── auth.go            # Authentication logic
│   │   │   └── presence.go        # User presence tracking
│   │   │
│   │   ├── repository/
│   │   │   ├── user.go            # User data access
│   │   │   └── redis.go           # Redis operations
│   │   │
│   │   ├── grpc/
│   │   │   └── server.go          # gRPC server implementation
│   │   │
│   │   ├── middleware/
│   │   │   ├── auth.go            # JWT validation
│   │   │   └── logging.go         # Request logging
│   │   │
│   │   └── models/
│   │       ├── user.go            # User domain model
│   │       └── session.go         # Session model
│   │
│   ├── pkg/
│   │   ├── jwt/
│   │   │   └── jwt.go             # JWT utilities
│   │   ├── logger/
│   │   │   └── logger.go
│   │   └── validator/
│   │       └── validator.go
│   │
│   ├── migrations/
│   │   ├── 000001_create_users_table.up.sql
│   │   ├── 000001_create_users_table.down.sql
│   │   ├── 000002_create_sessions_table.up.sql
│   │   └── 000002_create_sessions_table.down.sql
│   │
│   ├── pb/                        # Generated protobuf code
│   │   ├── user/
│   │   └── common/
│   │
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── .air.toml
│   └── .env
│
├── scripts/                       # Utility scripts
│   ├── generate-proto.sh          # Generate protobuf code
│   ├── migrate-up.sh              # Run migrations
│   ├── migrate-down.sh            # Rollback migrations
│   └── seed-data.sh               # Seed test data
│
└── docs/                          # Documentation
    ├── architecture.md
    ├── api/
    │   ├── chat-service.md
    │   └── user-service.md
    ├── grpc/
    │   └── services.md
    └── deployment.md


Key Files Content Examples:
════════════════════════════

# docker-compose.yml location: ./docker-compose.yml
# .env.example location: ./
# Makefile location: ./

# chat-service/cmd/main.go
# user-service/cmd/main.go

# proto/user/user.proto
# proto/chat/chat.proto

Directory Structure Highlights:
═══════════════════════════════

1. ROOT LEVEL:
   - docker-compose.yml: Orchestrates all services
   - proto/: Shared protobuf definitions
   - Makefile: Common commands (build, test, proto-gen)

2. CHAT-SERVICE:
   - internal/websocket/: Gorilla WebSocket management
   - internal/queue/: RabbitMQ publisher/consumer
   - internal/grpc/: gRPC client to call user service

3. USER-SERVICE:
   - internal/grpc/: gRPC server implementation
   - pkg/jwt/: JWT token generation & validation
   - internal/service/presence.go: User presence tracking

4. SHARED:
   - proto/: Protocol buffer definitions
   - pb/: Generated Go code (in each service)
   - migrations/: SQL migrations for each service

5. SUPPORTING:
   - scripts/: Build and deployment scripts
   - docs/: API and architecture documentation
