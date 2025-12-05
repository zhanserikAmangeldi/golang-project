module chat-service

go 1.24.1

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.3
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/zhanserikAmangeldi/proto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.67.0
)

require (
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace github.com/zhanserikAmangeldi/proto => ../proto
