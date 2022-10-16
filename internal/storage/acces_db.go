package storage

import (
	"fmt"
	"gophermart/internal/service"
)

func (dbStorage DBStorage) RegisterUser(user service.User) error {
	var dbUser service.User
	dbStorage.db.Where("login = ?", user.Login).First(&dbUser)

	//check if email is already registered
	if dbUser.Login != "" {
		return ErrUserExists
	}

	hashedPassword, err := service.GeneratePasswordHash(user.Password)
	if err != nil {
		return fmt.Errorf("error in password hashing: %s", err)
	}
	user.Password = hashedPassword
	//insert user details in database
	dbStorage.db.Create(&user)

	return nil
}

func (dbStorage DBStorage) CheckUserAuth(authDetails service.Authentication) error {
	var authUser service.User

	dbStorage.db.Where("login  = 	?", authDetails.Login).First(&authUser)
	if authUser.Login == "" {
		return ErrInvalidCredentials
	}

	if !service.CheckPasswordHash(authDetails.Password, authUser.Password) {
		return ErrInvalidCredentials
	}
	return nil
}

func (dbStorage DBStorage) PutOrder(order service.Order) error {
	var checkingOrder service.Order
	dbStorage.db.Where("login  = 	?  AND order_id = ?", order.Login, order.OrderID).First(&checkingOrder)
	if checkingOrder.Login != "" {
		return ErrAlreadyExists
	}
	dbStorage.db.Where("order_id = ?", order.OrderID).First(&checkingOrder)
	if checkingOrder.Login != "" {
		return ErrUploadedByAnotherUser
	}
	dbStorage.db.Create(&order)
	return nil
}
