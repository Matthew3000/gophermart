package app

import (
	"github.com/gorilla/mux"
	"gophermart/internal/config"
	"gophermart/internal/storage"
	"gophermart/internal/tools"
	"log"
	"net/http"
)

/*
   POST /api/user/register — регистрация пользователя;
   POST /api/user/login — аутентификация пользователя;
   POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
   GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
   GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
   POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
   GET /api/user/balance/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
*/

type App struct {
	config      config.Config
	userStorage storage.UserStorage
}

func NewApp(cfg config.Config, userStorage storage.UserStorage) *App {
	return &App{config: cfg, userStorage: userStorage}
}

func (app *App) Run() {
	router := mux.NewRouter()

	router.Use(tools.GzipMiddleware)

	router.HandleFunc("/api/user/register", app.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/api/user/login", app.handleLogin).Methods(http.MethodPost)
	router.HandleFunc("/api/user/orders", app.handleUploadOrder).Methods(http.MethodPost)
	router.HandleFunc("/api/user/orders", app.handleGetOrders).Methods(http.MethodGet)
	router.HandleFunc("/api/user/balance", app.handleGetBalance).Methods(http.MethodGet)
	router.HandleFunc("/api/user/balance/withdraw", app.handleWithdraw).Methods(http.MethodPost)
	router.HandleFunc("/api/user/balance/withdrawals", app.handleWithdrawInfo).Methods(http.MethodGet)

	router.HandleFunc("/", app.handleDefault)

	log.Fatal(http.ListenAndServe(app.config.ServerAddress, router))
}
