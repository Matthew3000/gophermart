package service

import (
	"time"
)

type Order struct {
	OrderID    string    `gorm:"unique" json:"number,omitempty"`
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
