package grpc

import (
	"context"
	"log"
	"time"

	pb "github.com/zhanserikAmangeldi/chat-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
}

// NewUserClient establishes the connection to the User Service.
func NewUserClient(address string) (*UserClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := pb.NewUserServiceClient(conn)

	return &UserClient{
		client: c,
		conn:   conn,
	}, nil
}

// Close cleans up the gRPC connection
func (c *UserClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// ValidateUserExists calls the User Service to check if a single user ID is valid.
func (c *UserClient) ValidateUserExists(ctx context.Context, userID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req := &pb.GetUserRequest{
		Id: userID,
	}

	_, err := c.client.GetUser(ctx, req)
	if err != nil {
		log.Printf("[gRPC] ValidateUserExists: ID %d not found or service down: %v", userID, err)
		return false, nil
	}

	return true, nil
}

// ValidateUsersExist checks a list of users (used when creating a group).
func (c *UserClient) ValidateUsersExist(ctx context.Context, userIDs []int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req := &pb.CheckUsersExistRequest{
		UserIds: userIDs,
	}

	res, err := c.client.CheckUsersExist(ctx, req)
	if err != nil {
		log.Printf("[gRPC] ValidateUsersExist error: %v", err)
		return false, err
	}

	return res.Exists, nil
}
