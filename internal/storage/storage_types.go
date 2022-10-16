package storage

import (
	"errors"
	"github.com/gorilla/sessions"
	"gophermart/internal/service"
)

type UserStorage interface {
	CheckUserAuth(authDetails service.Authentication) (service.Token, error)
	RegisterUser(user service.User) error
	PutOrder(order service.Order) error
}

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

var CookieStorage = sessions.NewCookieStore([]byte("secret_key"))
