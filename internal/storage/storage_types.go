package storage

import (
	"errors"
	"gophermart/internal/auth"
)

type UserStorage interface {
	CheckUserAuth(authDetails auth.Authentication) (auth.Token, error)
	RegisterUser(user auth.User) error
}

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
