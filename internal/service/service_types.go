package service

import (
	"github.com/jinzhu/gorm"
	"time"
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

type Balance struct {
	Login     string  `json:"-"`
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

//type Withdrawals struct {
//	Login       string
//	Amount      float32
//	ProcessedAt time.Time `json:"processed_at,omitempty"`
//}

type Withdrawal struct {
	Login       string    `json:"-"`
	OrderID     string    `json:"order"`
	Amount      float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}
