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

func (dbStorage DBStorage) CheckUserAuth(authDetails service.Authentication) (service.Token, error) {
	var authUser service.User
	var token service.Token

	dbStorage.db.Where("login  = 	?", authDetails.Login).First(&authUser)
	if authUser.Login == "" {
		return token, ErrInvalidCredentials
	}

	if !service.CheckPasswordHash(authDetails.Password, authUser.Password) {
		return token, ErrInvalidCredentials
	}

	validToken, err := service.GenerateJWT(authUser.Login)
	if err != nil {
		return token, fmt.Errorf("creating token: %s", err)
	}

	token.Login = authUser.Login
	token.TokenString = validToken
	return token, nil
}

func (dbStorage DBStorage) PutOrder(order service.Order) error {
	dbStorage.db.Create(&order)
	return nil
}
