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

	err := r.ParseForm()
	if err != nil {
		log.Print(err)
		http.Error(w, fmt.Sprintf("parse form error: %s", err), http.StatusBadRequest)
		return
	}
	user.Login = r.Form.Get("login")
	user.Password = r.Form.Get("password")

	err = auth.CheckLoginAndPassword(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("register error: %s", err), http.StatusConflict)
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
	w.WriteHeader(http.StatusCreated)
}

func (app *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var authDetails auth.Authentication

	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("parse form error: %s", err), http.StatusBadRequest)
		return
	}
	authDetails.Login = r.Form.Get("login")
	authDetails.Password = r.Form.Get("password")

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

}
