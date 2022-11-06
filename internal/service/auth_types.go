package service

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string  `json:"name"`
	Login    string  `json:"login" gorm:"unique"`
	Password string  `json:"password"`
	Balance  float32 `json:"accrual,omitempty"`
}

type Authentication struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

var SecretKey = "watch?v=Qw4w9WgXcQ"
