package storage

import (
	"errors"
	"gophermart/internal/service"
)

type UserStorage interface {
	CheckUserAuth(authDetails service.Authentication) error
	RegisterUser(user service.User) error
	PutOrder(order service.Order) error
}

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
