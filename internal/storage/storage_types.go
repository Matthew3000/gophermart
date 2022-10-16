package storage

import (
	"errors"
	"gophermart/internal/service"
)

type UserStorage interface {
	CheckUserAuth(authDetails service.Authentication) error
	RegisterUser(user service.User) error
	PutOrder(order service.Order, serverAddr string) error
	GetOrderStatus(order service.Order, serverAddr string) (service.Order, error)
}

var (
	ErrUserExists            = errors.New("user already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAlreadyExists         = errors.New("this order is already uploaded")
	ErrUploadedByAnotherUser = errors.New("this order is uploaded by another user")
)
