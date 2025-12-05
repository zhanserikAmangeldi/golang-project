package grpcserver

import (
	"context"
	
	userpb "github.com/zhanserikAmangeldi/proto/userpb"
	"github.com/zhanserikAmangeldi/user-service/internal/repository"
)

type UserGrpcServer struct {
	userpb.UnimplementedUserServiceServer
	userRepo *repository.UserRepository
}

func NewUserGrpcServer(repo *repository.UserRepository) *UserGrpcServer {
	return &UserGrpcServer{userRepo: repo}
}

func (s *UserGrpcServer) GetUserById(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	// Создаем User для ответа
	respUser := &userpb.User{
		Id:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Status:   user.Status,
	}

	// Проверяем и разыменовываем указатели
	if user.DisplayName != nil {
		respUser.DisplayName = *user.DisplayName
	}
	
	if user.AvatarURL != nil {
		respUser.AvatarUrl = *user.AvatarURL
	}

	// Возвращаем UserResponse
	return &userpb.UserResponse{
		User: respUser,
	}, nil
}