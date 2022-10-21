package storage

import (
	"errors"
	"gophermart/internal/service"
)

type UserStorage interface {
	CheckUserAuth(authDetails service.Authentication) error
	RegisterUser(user service.User) error
	PutOrder(order service.Order) error
	UpdateAccrual(accrualAddr string) error
	GetOrdersByLogin(login string) ([]service.Order, error)
	GetBalanceByLogin(login string) (float32, error)
	GetWithdrawnAmount(login string) (float32, error)
	Withdraw(withdrawal service.Withdrawal) error
	SetBalanceByLogin(login string, newBalance float32) error
	GetWithdrawals(login string) ([]service.Withdrawal, error)
}

var (
	ErrUserExists            = errors.New("user already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAlreadyExists         = errors.New("this order is already uploaded")
	ErrUploadedByAnotherUser = errors.New("this order is uploaded by another user")
	ErrOrderListEmpty        = errors.New("order list is empty")
	ErrWithdrawListEmpty     = errors.New("withdraw list is empty")
)