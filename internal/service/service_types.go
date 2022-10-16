package service

import (
	"github.com/jinzhu/gorm"
)

type Order struct {
	gorm.Model
	Login   string `gorm:"unique" json:"login"`
	OrderID int    `gorm:"unique" json:"order_id"`
}
