package mocks

import (
	"context"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/grpc"

	"github.com/stretchr/testify/mock"
)

type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) NewUserClient(address string) (*grpc.UserClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockUserClient) ValidateUserExists(ctx context.Context, userID int64) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserClient) ValidateUsersExist(ctx context.Context, userIDs []int64) (bool, error) {
	args := m.Called(ctx, userIDs)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserClient) Close() {
	m.Called()
}
