package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"gophermart/internal/service"
	"io"
	"log"
	"net/http"
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

func (dbStorage DBStorage) PutOrder(order service.Order, serverAddr string) error {
	var checkingOrder service.Order
	dbStorage.db.Where("login  = 	?  AND order_id = ?", order.Login, order.OrderID).First(&checkingOrder)
	if checkingOrder.Login != "" {
		return ErrAlreadyExists
	}
	dbStorage.db.Where("order_id = ?", order.OrderID).First(&checkingOrder)
	if checkingOrder.Login != "" {
		return ErrUploadedByAnotherUser
	}
	var err error
	order, err = dbStorage.GetOrderStatus(order, serverAddr)
	if err != nil {
		return err
	}
	dbStorage.db.Create(&order)
	return nil
}

func (dbStorage DBStorage) GetOrderStatus(order service.Order, serverAddr string) (service.Order, error) {
	getStatusURL := "http://localhost:36293/api/orders/" + order.OrderID
	log.Printf("get order status from: %s", getStatusURL)

	response, err := http.Get(getStatusURL)
	if err != nil {
		log.Printf("get order status: %s", err)
		return order, err
	}
	defer response.Body.Close()

	value, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("read request body err: %s", err)
		return order, err
	}
	log.Printf("request body is  %s", value)

	var orderResponse service.OrderAccrualResponse
	if err := json.NewDecoder(response.Body).Decode(&orderResponse); err != nil {
		log.Printf("json response body: %s", response.Body)
		log.Printf("json decode order accrual: %s", err)
		return order, err
	}
	if orderResponse.OrderID == "" {
		return order, errors.New("empty JSON")
	}
	order.Status = orderResponse.Status
	order.Accrual = orderResponse.Accrual
	log.Printf("order status: %s, order accrual: %d", order.Status, order.Accrual)
	return order, nil
}
