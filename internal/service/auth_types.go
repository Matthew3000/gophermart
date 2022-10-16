package service

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Name     string `json:"name"`
	Login    string `gorm:"unique" json:"login"`
	Password string `json:"password"`
}

type Authentication struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

var SecretKey = "watch?v=Qw4w9WgXcQ"
