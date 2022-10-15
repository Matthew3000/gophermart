package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"gophermart/internal/auth"
	"gophermart/internal/storage"
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

func (app *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	var user auth.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, fmt.Sprintf("parse form error: %s", err), http.StatusBadRequest)
		return
	}

	err = app.userStorage.RegisterUser(user)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			http.Error(w, fmt.Sprintf("register error: %s", err), http.StatusConflict)
			return
		} else {
			http.Error(w, fmt.Sprintf("register error: %s", err), http.StatusInternalServerError)
			return
		}
	}

	var authDetails auth.Authentication
	authDetails.Login = user.Login
	authDetails.Password = user.Password
	token, err := app.userStorage.CheckUserAuth(authDetails)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidCredentials) {
			log.Printf("user: %s, password: %s", authDetails.Login, authDetails.Password)
			http.Error(w, fmt.Sprintf("auth error: %s", err), http.StatusUnauthorized)
			return
		} else {
			http.Error(w, fmt.Sprintf("auth error: %s", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", "JWT realm=\"api\"")
	json.NewEncoder(w).Encode(token)
	w.WriteHeader(http.StatusOK)
}

func (app *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var authDetails auth.Authentication
	err := json.NewDecoder(r.Body).Decode(&authDetails)
	if err != nil {
		http.Error(w, fmt.Sprintf("parse form error: %s", err), http.StatusBadRequest)
		return
	}

	token, err := app.userStorage.CheckUserAuth(authDetails)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidCredentials) {
			http.Error(w, fmt.Sprintf("auth error: %s", err), http.StatusUnauthorized)
			return
		} else {
			http.Error(w, fmt.Sprintf("auth error: %s", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func (app *App) handleUploadOrder(w http.ResponseWriter, r *http.Request) {

}

func (app *App) handleGetOrders(w http.ResponseWriter, r *http.Request) {

}

func (app *App) handleGetBalance(w http.ResponseWriter, r *http.Request) {

}

func (app *App) handleWithdraw(w http.ResponseWriter, r *http.Request) {

}

func (app *App) handleWithdrawInfo(w http.ResponseWriter, r *http.Request) {

}

func (app *App) handleDefault(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", http.StatusTemporaryRedirect)
}
