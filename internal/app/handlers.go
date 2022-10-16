package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/theplant/luhn"
	"gophermart/internal/service"
	"gophermart/internal/storage"
	"io"
	"log"
	"net/http"
	"strconv"
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

func (app *App) IsAuthorized(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := app.cookieStorage.Get(r, "session.id")
		authenticated := session.Values["authenticated"]
		if authenticated != nil && authenticated != false {
			handler.ServeHTTP(w, r)
		}
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
	}
}

func (app *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	var user service.User
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

	var authDetails service.Authentication
	authDetails.Login = user.Login
	authDetails.Password = user.Password
	err = app.userStorage.CheckUserAuth(authDetails)
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

	session, _ := app.cookieStorage.Get(r, "session.id")
	session.Values["authenticated"] = true
	session.Save(r, w)
	w.WriteHeader(http.StatusOK)
}

func (app *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var authDetails service.Authentication
	err := json.NewDecoder(r.Body).Decode(&authDetails)
	if err != nil {
		http.Error(w, fmt.Sprintf("parse form error: %s", err), http.StatusBadRequest)
		return
	}

	err = app.userStorage.CheckUserAuth(authDetails)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidCredentials) {
			http.Error(w, fmt.Sprintf("auth error: %s", err), http.StatusUnauthorized)
			return
		} else {
			http.Error(w, fmt.Sprintf("auth error: %s", err), http.StatusInternalServerError)
			return
		}
	}

	session, _ := app.cookieStorage.Get(r, "session.id")
	session.Values["authenticated"] = true
	session.Save(r, w)
	w.WriteHeader(http.StatusOK)
}

func (app *App) handleUploadOrder(w http.ResponseWriter, r *http.Request) {
	value, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("handle Upload Order: read request body: %v\n", err)
		http.Error(w, "couldn't read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderID, _ := strconv.Atoi(string(value))
	if !luhn.Valid(orderID) {
		log.Printf("handle Upload Order: order number is invalid")
		http.Error(w, "order number is invalid", http.StatusUnprocessableEntity)
		return
	}

	var order service.Order
	order.OrderID = orderID
	//order.Login = user.Login
	err = app.userStorage.PutOrder(order)
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusAccepted)
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
