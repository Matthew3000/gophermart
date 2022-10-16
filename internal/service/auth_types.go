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

type Token struct {
	Login       string `json:"login"`
	TokenString string `json:"token"`
}

var secretKey = "watch?v=Qw4w9WgXcQ"
