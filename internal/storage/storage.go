package storage

import (
	"context"
	"errors"
	"gophermart/internal/service"
	"time"
)

type UserStorage interface {
	CheckUserAuth(authDetails service.Authentication, ctx context.Context) error
	RegisterUser(user service.User, ctx context.Context) error
	PutOrder(order service.Order, ctx context.Context) error
	GetOrdersByLogin(login string, ctx context.Context) ([]service.Order, error)
	GetBalanceByLogin(login string, ctx context.Context) (float32, error)
	GetWithdrawnAmount(login string, ctx context.Context) (float32, error)
	Withdraw(withdrawal service.Withdrawal, ctx context.Context) error
	GetWithdrawals(login string, ctx context.Context) ([]service.Withdrawal, error)
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
	ErrNotEnoughPoints       = errors.New("not enough accural points")
)

const (
	NEW                = "NEW"
	REGISTERED         = "REGISTERED"
	PROCESSING         = "PROCESSING"
	UPDATEACCURALTIMER = 5 * time.Second
)
