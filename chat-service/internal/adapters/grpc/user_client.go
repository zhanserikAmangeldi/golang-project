package grpc

import (
	"context"
	"log"
	"time"

	pb "github.com/zhanserikAmangeldi/chat-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type IUserClient interface {
	NewUserClient(address string) (*UserClient, error)
	ValidateUserExists(ctx context.Context, userID int64) (bool, error)
	ValidateUsersExist(ctx context.Context, userIDs []int64) (bool, error)
	Close()
}

type UserClient struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
}

func (c *UserClient) NewUserClient(address string) (*UserClient, error) {
	//TODO implement me
	panic("implement me")
}

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

func (c *UserClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *UserClient) ValidateUserExists(ctx context.Context, userID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	log.Printf("[gRPC] ValidateUsersExist success %v\n", userID)

	req := &pb.GetUserRequest{
		Id: userID,
	}

	res, err := c.client.GetUser(ctx, req)
	if err != nil {
		log.Printf("[gRPC] ValidateUserExists: ID %d not found or service down: %v", userID, err)
		return false, nil
	}
	log.Printf("[gRPC] ValidateUsersExist success %v\n", res)

	return true, nil
}

func (c *UserClient) ValidateUsersExist(ctx context.Context, userIDs []int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	log.Printf("[gRPC] ValidateUsersExist success %v\n", userIDs)

	req := &pb.CheckUsersExistRequest{
		UserIds: userIDs,
	}

	res, err := c.client.CheckUsersExist(ctx, req)
	if err != nil {
		log.Printf("[gRPC] ValidateUsersExist error: %v", err)
		return false, err
	}
	log.Printf("[gRPC] ValidateUsersExist success %v\n", res)
	return res.Exists, nil
}
