package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	app "gophermart/internal"
	"gophermart/internal/config"
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
	flag.Parse()

	fmt.Println(cfg.DatabaseDSN)

	userStorage := storage.NewUserStorage(cfg.DatabaseDSN)
	var application = app.NewApp(cfg, userStorage)
	application.Run()
}
