package storage

import (
	"errors"
	"gophermart/internal/service"
	"time"
)

type UserStorage interface {
	CheckUserAuth(authDetails service.Authentication) error
	RegisterUser(user service.User) error
	//PutOrder(order service.Order, ctx context.Context) error
	PutOrder(order service.Order) error
	GetOrdersByLogin(login string) ([]service.Order, error)
	GetBalanceByLogin(login string) (float32, error)
	GetWithdrawnAmount(login string) (float32, error)
	Withdraw(withdrawal service.Withdrawal) error
	SetBalanceByLogin(login string, newBalance float32) error
	GetWithdrawals(login string) ([]service.Withdrawal, error)
	GetOrdersToUpdate() ([]service.Order, error)
	UpdateOrderStatus(order service.Order) error
	DeleteAll()
}

var (
	ErrUserExists            = errors.New("user already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAlreadyExists         = errors.New("this order is already uploaded")
	ErrUploadedByAnotherUser = errors.New("this order is uploaded by another user")
	ErrOrderListEmpty        = errors.New("order list is empty")
	ErrWithdrawListEmpty     = errors.New("withdraw list is empty")
)

const (
	NEW                = "NEW"
	REGISTERED         = "REGISTERED"
	PROCESSING         = "PROCESSING"
	UPDATEACCURALTIMER = 5 * time.Second
)
