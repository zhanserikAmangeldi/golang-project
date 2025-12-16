package grpc

import (
	"context"

	"github.com/zhanserikAmangeldi/user-service/internal/repository"
	pb "github.com/zhanserikAmangeldi/user-service/proto"
)

type UserGrpcServer struct {
	pb.UnimplementedUserServiceServer
	repo *repository.UserRepository
}

func NewUserGrpcServer(repo *repository.UserRepository) *UserGrpcServer {
	return &UserGrpcServer{repo: repo}
}

// GetUser returns details for a single user
func (s *UserGrpcServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	avatar := ""
	if user.AvatarURL != nil {
		avatar = *user.AvatarURL
	}

	return &pb.GetUserResponse{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarUrl: avatar,
	}, nil
}

// CheckUsersExist checks if a list of users exists (used for Groups)
func (s *UserGrpcServer) CheckUsersExist(ctx context.Context, req *pb.CheckUsersExistRequest) (*pb.CheckUsersExistResponse, error) {
	for _, id := range req.UserIds {
		_, err := s.repo.GetByID(ctx, id)
		if err != nil {
			// If even one user is missing, return false
			return &pb.CheckUsersExistResponse{Exists: false}, nil
		}
	}

	return &pb.CheckUsersExistResponse{Exists: true}, nil
}
