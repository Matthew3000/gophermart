package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/render"
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
			return
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
	session.Values["login"] = user.Login
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
	session.Values["login"] = authDetails.Login
	session.Save(r, w)
	w.WriteHeader(http.StatusOK)
}

func (app *App) handleUploadOrder(w http.ResponseWriter, r *http.Request) {
	value, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("handle Upload Order: read request body: %s", err)
		http.Error(w, "couldn't read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderID, _ := strconv.Atoi(string(value))
	if !luhn.Valid(orderID) {
		log.Printf("handle upload order: order number is invalid")
		http.Error(w, "order number is invalid", http.StatusUnprocessableEntity)
		return
	}

	var order service.Order
	order.OrderID = fmt.Sprint(orderID)

	session, _ := app.cookieStorage.Get(r, "session.id")
	order.Login = session.Values["login"].(string)
	log.Printf("%s", order.Login)
	err = app.userStorage.PutOrder(order)
	if err != nil {
		log.Printf("handle upload order: %s", err)
		if errors.Is(err, storage.ErrAlreadyExists) {
			http.Error(w, fmt.Sprint(err), http.StatusOK)
		} else if errors.Is(err, storage.ErrUploadedByAnotherUser) {
			http.Error(w, fmt.Sprint(err), http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprint(err), http.StatusUnprocessableEntity)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

func (app *App) handleGetOrders(w http.ResponseWriter, r *http.Request) {
	var order service.Order

	session, _ := app.cookieStorage.Get(r, "session.id")
	order.Login = session.Values["login"].(string)

	listOrders, err := app.userStorage.GetOrdersByLogin(order.Login)
	//log.Print(listOrders)
	if err != nil {
		if errors.Is(err, storage.ErrOrderListEmpty) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			log.Printf("handle get orders: %s", err)
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
	}
	render.JSON(w, r, listOrders)
}

func (app *App) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	var balance service.Balance
	session, _ := app.cookieStorage.Get(r, "session.id")
	balance.Login = session.Values["login"].(string)

	currentBalance, err := app.userStorage.GetBalanceByLogin(balance.Login)
	if err != nil {
		log.Printf("handle withdraw: get balance: %s", err)
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
	}

	balance.Current = currentBalance
	withdrawn, err := app.userStorage.GetWithdrawnAmount(balance.Login)
	if err != nil {
		log.Printf("handle withdraw: get withdrawn amount: %s", err)
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
	}

	balance.Withdrawn = withdrawn
	render.JSON(w, r, balance)
}

func (app *App) handleWithdraw(w http.ResponseWriter, r *http.Request) {
	var withdrawal service.Withdrawal

	session, _ := app.cookieStorage.Get(r, "session.id")
	withdrawal.Login = session.Values["login"].(string)

	err := json.NewDecoder(r.Body).Decode(&withdrawal)
	if err != nil {
		log.Printf("handle withdraw: read request body: %s", err)
		http.Error(w, "couldn't read body", http.StatusBadRequest)
		return
	}

	orderID, _ := strconv.Atoi(withdrawal.OrderID)
	if !luhn.Valid(orderID) {
		log.Printf("handle upload order: order number is invalid")
		http.Error(w, "order number is invalid", http.StatusUnprocessableEntity)
		return
	}

	currentBalance, err := app.userStorage.GetBalanceByLogin(withdrawal.Login)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	newBalance := currentBalance - withdrawal.Amount
	if newBalance >= 0 {
		err = app.userStorage.Withdraw(withdrawal)
		if err != nil {
			log.Printf("handle withdraw: withdraw: %s", err)
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		err = app.userStorage.SetBalanceByLogin(withdrawal.Login, newBalance)
		if err != nil {
			log.Printf("handle withdraw: set balance: %s", err)
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
}

func (app *App) handleWithdrawInfo(w http.ResponseWriter, r *http.Request) {

	session, _ := app.cookieStorage.Get(r, "session.id")
	login := session.Values["login"].(string)

	listWithdrawals, err := app.userStorage.GetWithdrawals(login)
	log.Print(listWithdrawals)
	if err != nil {
		log.Printf("handle withdraw info: get withdarw: %s", err)
		if errors.Is(err, storage.ErrWithdrawListEmpty) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err)
		}
	}

	render.JSON(w, r, listWithdrawals)
}

func (app *App) handleDefault(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", http.StatusTemporaryRedirect)
}
