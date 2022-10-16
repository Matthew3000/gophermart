package service

import (
	"github.com/jinzhu/gorm"
)

type Order struct {
	gorm.Model
	Login   string `gorm:"unique" json:"login"`
	OrderID string `gorm:"unique" json:"order_id"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual"`
}

type OrderAccrualResponse struct {
	OrderID string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual"`
}
