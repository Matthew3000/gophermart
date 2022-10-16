package storage

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gophermart/internal/service"
	"log"
)

type DBStorage struct {
	db *gorm.DB
}

func NewUserStorage(DatabaseURL string) *DBStorage {
	connection, err := gorm.Open("postgres", DatabaseURL)
	if err != nil {
		log.Fatalf("database failed to open: %s", err)
	}
	sqlDB := connection.DB()

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("database failed to ping: %s", err)
	}
	log.Printf("Database connection successful")

	InitializeTables(connection)

	return &DBStorage{
		db: connection,
	}
}

func InitializeTables(connection *gorm.DB) {
	connection.AutoMigrate(service.User{})
	connection.AutoMigrate(service.Order{})
}
