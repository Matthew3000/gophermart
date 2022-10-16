package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/gorilla/sessions"
	"gophermart/internal/app"
	"gophermart/internal/config"
	"gophermart/internal/service"
	"gophermart/internal/storage"
	"log"
)

func main() {
	var cfg config.Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "File Storage Path")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Server address")
	flag.StringVar(&cfg.AccrualAddress, "r", cfg.AccrualAddress, "Accrual address")
	flag.Parse()

	fmt.Println(cfg.DatabaseDSN)

	userStorage := storage.NewUserStorage(cfg.DatabaseDSN)
	cookieStorage := sessions.NewCookieStore([]byte(service.SecretKey))
	var application = app.NewApp(cfg, userStorage, *cookieStorage)
	application.Run()
}
