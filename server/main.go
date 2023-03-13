package demo

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	// импортируем пакет со сгенерированными protobuf-файлами
	pb "demo/proto"
)

// UsersServer поддерживает все необходимые методы сервера.
type UsersServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedUsersServer

	// используем sync.Map для хранения пользователей
	users sync.Map
}

// AddUser реализует интерфейс добавления пользователя.
func (s *UsersServer) AddUser(ctx context.Context, in *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	var response pb.AddUserResponse

	if _, ok := s.users.Load(in.User.Email); ok {
		response.Error = fmt.Sprintf("Пользователь с email %s уже существует", in.User.Email)
	} else {
		s.users.Store(in.User.Email, in.User)
	}
	return &response, nil
}

// ListUsers реализует интерфейс получения списка пользователей.
func (s *UsersServer) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	var list []string

	s.users.Range(func(key, _ interface{}) bool {
		list = append(list, key.(string))
		return true
	})
	// сортируем слайс из email
	sort.Strings(list)

	offset := int(in.Offset)
	end := int(in.Offset + in.Limit)
	if end > len(list) {
		end = len(list)
	}
	if offset >= end {
		offset = 0
		end = 0
	}
	response := pb.ListUsersResponse{
		Count:  int32(len(list)),
		Emails: list[offset:end],
	}
	return &response, nil
}

func (s *UsersServer) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	var response pb.GetUserResponse
	var err error
	user, ok := s.users.Load(in.Email)
	if ok {
		response.User = user.(*pb.User)
	} else {
		response.Error = "user not found"
		err = errors.New("user not found")
	}
	return &response, err
}

func (s *UsersServer) DelUser(ctx context.Context, in *pb.DelUserRequest) (*pb.DelUserResponse, error) {
	var response pb.DelUserResponse
	var err error

	_, ok := s.users.LoadAndDelete(in.Email)
	if !ok {
		response.Error = "user not found"
		err = errors.New("user not found")
	}

	return &response, err
}
