package storage

import (
	"context"
	"errors"
	"fmt"
	"gophermart/internal/service"
	"gorm.io/gorm"
	"time"
)

func (dbStorage DBStorage) RegisterUser(user service.User, ctx context.Context) error {
	var dbUser service.User
	err := dbStorage.db.WithContext(ctx).Where("login = ?", user.Login).First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			hashedPassword, err := service.GeneratePasswordHash(user.Password)
			if err != nil {
				return fmt.Errorf("error in password hashing: %s", err)
			}
			user.Password = hashedPassword
			user.Balance = 0
			err = dbStorage.db.WithContext(ctx).Create(&user).Error
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return ErrUserExists
}

func (dbStorage DBStorage) CheckUserAuth(authDetails service.Authentication, ctx context.Context) error {
	var authUser service.User

	err := dbStorage.db.WithContext(ctx).Where("login  = 	?", authDetails.Login).First(&authUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidCredentials
		}
		return err
	}

	if !service.CheckPasswordHash(authDetails.Password, authUser.Password) {
		return ErrInvalidCredentials
	}
	return nil
}

func (dbStorage DBStorage) PutOrder(order service.Order, ctx context.Context) error {
	var checkingOrder service.Order

	err := dbStorage.db.WithContext(ctx).Where("login  = 	?  AND number = ?", order.Login, order.Number).First(&checkingOrder).Error
	if checkingOrder.Login != "" {
		return ErrAlreadyExists
	}
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	err = dbStorage.db.WithContext(ctx).Where("number = ?", order.Number).First(&checkingOrder).Error
	if checkingOrder.Login != "" {
		return ErrUploadedByAnotherUser
	}
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	order.Status = NEW
	order.UploadedAt = time.Now()
	err = dbStorage.db.WithContext(ctx).Create(&order).Error
	if err != nil {
		return err
	}

	return nil
}

func (dbStorage DBStorage) GetOrdersToUpdate() ([]service.Order, error) {
	var ordersToUpdate []service.Order
	err := dbStorage.db.Where("status = ?", NEW).Or("status = ?", REGISTERED).
		Or("status = ?", PROCESSING).Find(&ordersToUpdate).Error
	if err != nil {
		return nil, err
	}
	return ordersToUpdate, nil
}

func (dbStorage DBStorage) UpdateOrderStatus(order service.Order) error {
	err := dbStorage.db.Model(&service.Order{}).Where("number = ?", order.Number).
		Updates(service.Order{Status: order.Status, Accrual: order.Accrual}).Error
	if err != nil {
		return err
	}

	var user service.User
	err = dbStorage.db.Where("login  = 	?", order.Login).First(&user).Error
	if err != nil {
		return err
	}

	user.Balance = user.Balance + order.Accrual
	err = dbStorage.db.Save(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func (dbStorage DBStorage) GetOrdersByLogin(login string, ctx context.Context) ([]service.Order, error) {
	var orders []service.Order

	err := dbStorage.db.WithContext(ctx).Where("login  = 	?", login).Order("uploaded_at asc").Find(&orders).Error
	if len(orders) == 0 {
		return nil, ErrOrderListEmpty
	}
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (dbStorage DBStorage) GetBalanceByLogin(login string, ctx context.Context) (float32, error) {
	var user service.User
	err := dbStorage.db.WithContext(ctx).Where("login  = 	?", login).First(&user).Error
	if err != nil {
		return 0, err
	}
	return user.Balance, nil
}

func (dbStorage DBStorage) Withdraw(withdrawal service.Withdrawal, ctx context.Context) error {
	var user service.User
	err := dbStorage.db.WithContext(ctx).Where("login  = 	?", withdrawal.Login).First(&user).Error
	if err != nil {
		return err
	}

	newBalance := user.Balance - withdrawal.Amount
	if newBalance >= 0 {
		withdrawal.ProcessedAt = time.Now()
		err = dbStorage.db.WithContext(ctx).Save(&withdrawal).Where("order = ?", withdrawal.OrderID).Error
		if err != nil {
			return err
		}
		err = dbStorage.db.WithContext(ctx).Model(&service.User{}).
			Where("login = ?", withdrawal.Login).Update("balance", newBalance).Error
		if err != nil {
			return err
		}
		return nil
	}
	return ErrNotEnoughPoints
}

func (dbStorage DBStorage) GetWithdrawnAmount(login string, ctx context.Context) (float32, error) {
	var withdrawals []service.Withdrawal
	err := dbStorage.db.WithContext(ctx).Where("login  = 	?", login).Find(&withdrawals).Error
	if err != nil {
		return 0, err
	}

	var withdrawn float32
	for _, withdrawal := range withdrawals {
		withdrawn += withdrawal.Amount
	}

	return withdrawn, nil
}

func (dbStorage DBStorage) GetWithdrawals(login string, ctx context.Context) ([]service.Withdrawal, error) {
	var withdrawals []service.Withdrawal
	err := dbStorage.db.WithContext(ctx).Where("login  = 	?", login).Order("processed_at asc").Find(&withdrawals).Error
	if len(withdrawals) == 0 {
		return nil, ErrWithdrawListEmpty
	}
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (dbStorage DBStorage) DeleteAll(ctx context.Context) {
	dbStorage.db.WithContext(ctx).Exec("DELETE FROM users")
	dbStorage.db.WithContext(ctx).Exec("DELETE FROM orders")
}
