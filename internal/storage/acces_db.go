package storage

import (
	"fmt"
	"gophermart/internal/auth"
)

func (dbStorage DBStorage) RegisterUser(user auth.User) error {
	var dbUser auth.User
	dbStorage.db.Where("login = ?", user.Login).First(&dbUser)

	//check if email is already registered
	if dbUser.Login != "" {
		return ErrUserExists
	}

	hashedPassword, err := auth.GeneratePasswordHash(user.Password)
	if err != nil {
		return fmt.Errorf("error in password hashing: %s", err)
	}
	user.Password = hashedPassword
	//insert user details in database
	dbStorage.db.Create(&user)

	return nil
}

func (dbStorage DBStorage) CheckUserAuth(authDetails auth.Authentication) (auth.Token, error) {
	var authUser auth.User
	var token auth.Token

	dbStorage.db.Where("login  = 	?", authDetails.Login).First(&authUser)
	if authUser.Login == "" {
		return token, ErrInvalidCredentials
	}

	if !auth.CheckPasswordHash(authDetails.Password, authUser.Password) {
		return token, ErrInvalidCredentials
	}

	validToken, err := auth.GenerateJWT(authUser.Login)
	if err != nil {
		return token, fmt.Errorf("creating token: %s", err)
	}

	token.Login = authUser.Login
	token.TokenString = validToken
	return token, nil
}
