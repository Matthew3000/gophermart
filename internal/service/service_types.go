package service

import (
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
)

type Order struct {
	gorm.Model
	Login   string `gorm:"unique" json:"login"`
	OrderID int    `gorm:"unique" json:"order_id"`
}

var CookieStorage = sessions.NewCookieStore([]byte("secret_key"))
