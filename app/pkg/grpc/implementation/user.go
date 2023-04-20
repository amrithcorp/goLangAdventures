package implementation

import (
	"context"
	"core/core/pkg/grpc/protobuf"
	"core/core/pkg/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *Server) CreateUser(ctx context.Context, in *protobuf.CreateUserRequest) (*protobuf.CreateUserResponse, error) {
	UUID := uuid.New()
	createUser := s.Database.Create(&models.Account{
		ID:       UUID,
		Username: in.Username,
		Email:    in.Email,
	})
	if createUser.Error != nil {
		if errors.Is(createUser.Error, gorm.ErrDuplicatedKey) {
			return nil, errors.New("Email or Username already in use")
		} else {
			return nil, errors.New("")
		}
	}
	return &protobuf.CreateUserResponse{
		Id: UUID.String(),
	}, nil
}

func (s *Server) GetUser(ctx context.Context, in *protobuf.GetUserRequest) (*protobuf.GetUserResponse, error) {
	var account models.Account
	id, err := uuid.Parse(in.Id)
	if err != nil {
		return nil, errors.New("Invalid UUID")
	}
	searchUser := s.Database.First(&account, "id = ?", id)
	if errors.Is(searchUser.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("No User with this ID")
	}
	return &protobuf.GetUserResponse{
		Email:    account.Email,
		Username: account.Username,
	}, nil
}
