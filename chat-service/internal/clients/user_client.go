package clients

import (
	"context"
	"fmt"
	"time"
	
	"github.com/zhanserikAmangeldi/proto/userpb"
)

type UserClient interface {
	GetUser(ctx context.Context, id int64) (*userpb.User, error)
}

type userClient struct {
	grpc userpb.UserServiceClient
}

func NewUserClient(grpcClient userpb.UserServiceClient) UserClient {
	return &userClient{grpc: grpcClient}
}

func (c *userClient) GetUser(ctx context.Context, id int64) (*userpb.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Исправлено: GetUserById вместо GetUser, Id вместо UserId
	resp, err := c.grpc.GetUserById(ctx, &userpb.GetUserRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("user service GetUser error: %w", err)
	}

	if resp.GetUser() == nil {
		return nil, fmt.Errorf("user %d not found", id)
	}

	return resp.GetUser(), nil
}