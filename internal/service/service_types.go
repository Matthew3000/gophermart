package service

import (
	"github.com/jinzhu/gorm"
)

type Order struct {
	gorm.Model
	Login   string  `gorm:"unique" json:"login"`
	OrderID string  `gorm:"unique" json:"order_id,omitempty"`
	Status  string  `json:"status,omitempty"`
	Accrual float32 `json:"accrual,omitempty"`
}

type AccrualResponse struct {
	OrderID string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}
