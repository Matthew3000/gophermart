package app

import (
	"context"
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
	"time"
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

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		r = r.WithContext(ctx)

		http.Redirect(w, r, "/login", http.StatusUnauthorized)
	}
}

func (app *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	var user service.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("register err: json parse error: %s", err)
		http.Error(w, fmt.Sprintf("json parse error: %s", err), http.StatusBadRequest)
		return
	}

	err = app.userStorage.RegisterUser(user)
	if err != nil {
		log.Printf("register err: %s for user: %s, password: %s", err, user.Login, user.Password)
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
		log.Printf("register then auth err: %s for user: %s, password: %s", err, authDetails.Login, authDetails.Password)
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
	session.Values["login"] = user.Login
	session.Save(r, w)
	w.WriteHeader(http.StatusOK)
}

func (app *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var authDetails service.Authentication
	err := json.NewDecoder(r.Body).Decode(&authDetails)
	if err != nil {
		log.Printf("auth err: json parse error: %s", err)
		http.Error(w, fmt.Sprintf("json parse error: %s", err), http.StatusBadRequest)
		return
	}

	err = app.userStorage.CheckUserAuth(authDetails)
	if err != nil {
		log.Printf("auth err: %s for user: %s, password: %s", err, authDetails.Login, authDetails.Password)
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
		log.Printf("upload order: json parse error: %s", err)
		http.Error(w, "json parse error", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderID, _ := strconv.Atoi(string(value))
	if !luhn.Valid(orderID) {
		log.Printf("upload order: order number %s is invalid", string(value))
		http.Error(w, "order number is invalid", http.StatusUnprocessableEntity)
		return
	}

	var order service.Order
	order.Number = fmt.Sprint(orderID)

	session, _ := app.cookieStorage.Get(r, "session.id")
	order.Login = session.Values["login"].(string)
	err = app.userStorage.PutOrder(order, r.Context())
	if err != nil {
		log.Printf("upload order: %s for user: %s, number: %s", err, order.Login, order.Number)
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
	if err != nil {
		log.Printf("get orders: %s for user: %s", err, order.Login)
		if errors.Is(err, storage.ErrOrderListEmpty) {
			w.WriteHeader(http.StatusNoContent)
		} else {
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
		log.Printf("handle withdraw: get balance: %s for user: %s", err, balance.Login)
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
	}

	balance.Current = currentBalance
	withdrawn, err := app.userStorage.GetWithdrawnAmount(balance.Login)
	if err != nil {
		log.Printf("handle withdraw: get withdrawn amount: %s for user: %s", err, balance.Login)
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
		log.Printf("withdraw: json parse error: %s", err)
		http.Error(w, "json parse error:", http.StatusBadRequest)
		return
	}

	orderID, _ := strconv.Atoi(withdrawal.OrderID)
	if !luhn.Valid(orderID) {
		log.Printf("withdraw: order number %s is invalid", withdrawal.OrderID)
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
			log.Printf("withdraw: read request body: %s for user: %s amount: %f order: %s",
				err, withdrawal.Login, withdrawal.Amount, withdrawal.OrderID)
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		err = app.userStorage.SetBalanceByLogin(withdrawal.Login, newBalance)
		if err != nil {
			log.Printf("withdraw: read request body: %s for user: %s balance: %f order: %s",
				err, withdrawal.Login, newBalance, withdrawal.OrderID)
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
	if err != nil {
		log.Printf("withdraw info: get withdarw: %s for user: %s", err, login)
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
