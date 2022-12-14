package service

import (
	"time"
)

type Order struct {
	Number     string    `json:"number,omitempty" gorm:"unique"`
	Login      string    `json:"-"`
	Status     string    `json:"status,omitempty"`
	Accrual    float32   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at,omitempty"`
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

type Withdrawal struct {
	Login       string    `json:"-"`
	OrderID     string    `json:"order" gorm:"primaryKey"`
	Amount      float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}
