package storage

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"gophermart/internal/service"
	"log"
	"net/http"
	"time"
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
	user.Balance = 0
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

	order.Status = "NEW"
	order.UploadedAt = time.Now()
	dbStorage.db.Create(&order)
	return nil
}

func (dbStorage DBStorage) UpdateAccrual(accrualAddr string) error {
	var ordersToUpdate []service.Order
	dbStorage.db.Where("status = ?", "NEW").Or("status = ?", "REGISTERED").Or("status = ?", "PROCESSING").Find(&ordersToUpdate)

	if len(ordersToUpdate) != 0 {
		req := resty.New().
			SetBaseURL(accrualAddr).
			R().
			SetHeader("Content-Type", "application/json")
		for _, order := range ordersToUpdate {

			orderNum := order.OrderID
			resp, err := req.Get("/api/orders/" + orderNum)
			if err != nil {
				return err
			}

			status := resp.StatusCode()
			switch status {
			case http.StatusTooManyRequests:
				time.Sleep(10 * time.Second)
				return nil

			case http.StatusOK:
				var updatedOrder service.AccrualResponse
				err = json.Unmarshal(resp.Body(), &updatedOrder)
				if err != nil {
					log.Printf("json decode order accrual: %s", err)
					return err
				}
				log.Printf("accrual for order %s updating to %s", updatedOrder.OrderID, updatedOrder.Status)

				//order.Status = updatedOrder.Status
				//order.OrderID = updatedOrder.OrderID
				//order.Accrual = updatedOrder.Accrual
				//dbStorage.db.Save(&order)
				dbStorage.db.Model(&service.Order{}).Where("orderID = ?", updatedOrder.OrderID).Updates(service.Order{Status: updatedOrder.Status, Accrual: updatedOrder.Accrual})

				var user service.User
				dbStorage.db.Where("login  = 	?", order.Login).First(&user)
				user.Balance = user.Balance + updatedOrder.Accrual
				err := dbStorage.db.Save(&user).Error
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (dbStorage DBStorage) GetOrdersByLogin(login string) ([]service.Order, error) {
	var orders []service.Order

	dbStorage.db.Where("login  = 	?", login).Order("uploaded_at asc").Find(&orders)
	if len(orders) == 0 {
		return nil, ErrOrderListEmpty
	}

	return orders, nil
}

func (dbStorage DBStorage) GetBalanceByLogin(login string) (float32, error) {
	var user service.User
	err := dbStorage.db.Where("login  = 	?", login).First(&user).Error
	if err != nil {
		return 0, err
	}

	return user.Balance, nil
}

func (dbStorage DBStorage) GetWithdrawnAmount(login string) (float32, error) {
	var withdrawals []service.Withdrawal
	err := dbStorage.db.Where("login  = 	?", login).Find(&withdrawals).Error
	if err != nil {
		return 0, err
	}

	var withdrawn float32
	for _, withdrawal := range withdrawals {
		withdrawn += withdrawal.Amount
	}

	return withdrawn, nil
}

func (dbStorage DBStorage) Withdraw(withdrawal service.Withdrawal) error {
	withdrawal.ProcessedAt = time.Now()
	err := dbStorage.db.Save(&withdrawal).Error
	if err != nil {
		return err
	}
	return nil
}

func (dbStorage DBStorage) SetBalanceByLogin(login string, newBalance float32) error {
	err := dbStorage.db.Model(&service.User{}).Where("login = ?", login).Update("balance", newBalance).Error
	if err != nil {
		return err
	}
	return nil
}

func (dbStorage DBStorage) GetWithdrawals(login string) ([]service.Withdrawal, error) {
	var withdrawals []service.Withdrawal
	err := dbStorage.db.Where("login  = 	?", login).Order("processed_at asc").Find(&withdrawals).Error
	// .Order("created_at asc")
	if err != nil {
		return nil, err
	}
	if len(withdrawals) == 0 {
		return nil, ErrWithdrawListEmpty
	}

	return withdrawals, nil
}
