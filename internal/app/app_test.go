package app

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gophermart/internal/config"
	"gophermart/internal/service"
	"gophermart/internal/storage"
	"net/http"
	"testing"
)

func TestApp(t *testing.T) {

	cfg := config.Config{
		ServerAddress:  "localhost:8080",
		DatabaseDSN:    "postgres://matt:pvtjoker@localhost:5432/gophermart?sslmode=disable",
		AccrualAddress: "localhost:8080",
	}

	userStorage := storage.NewUserStorage(cfg.DatabaseDSN)
	cookieStorage := sessions.NewCookieStore([]byte(service.SecretKey))
	var app = NewApp(cfg, userStorage, *cookieStorage)
	go app.Run()

	cookie := RegisterTest(t, app)
	PutOrderTest(t, app, cookie)
	GetOrdersTest(t, app, cookie)
	GetBalanceTest(t, app, cookie)
	WithdrawTest(t, app, cookie)
	GetWithdrawalsTest(t, app, cookie)

	app.userStorage.DeleteAll()
}

func RegisterTest(t *testing.T, app *App) http.Cookie {
	var cookie http.Cookie
	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name   string
		addr   string
		method string
		user   service.User
		want   want
	}{
		{
			name:   "register ok",
			addr:   "/api/user/register",
			method: http.MethodPost,
			user: service.User{
				Login:    "nevergonna",
				Password: "giveyouup",
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "register conflict",
			addr:   "/api/user/register",
			method: http.MethodPost,
			user: service.User{
				Login:    "nevergonna",
				Password: "giveyouup",
			},
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "login fail: no such user",
			addr:   "/api/user/login",
			method: http.MethodPost,
			user: service.User{
				Login:    "imgona",
				Password: "giveyouup",
			},
			want: want{
				statusCode:  http.StatusUnauthorized,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "login fail: wrong pass",
			addr:   "/api/user/login",
			method: http.MethodPost,
			user: service.User{
				Login:    "nevergonna",
				Password: "letyoudown",
			},
			want: want{
				statusCode:  http.StatusUnauthorized,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "login OK",
			addr:   "/api/user/login",
			method: http.MethodPost,
			user: service.User{
				Login:    "nevergonna",
				Password: "giveyouup",
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.user)
			request := resty.New().R().SetHeader("Content-Type", "application/json").SetBody(body)

			result, err := request.Post("http://" + app.config.ServerAddress + tt.addr)
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode())
			assert.Equal(t, tt.want.contentType, result.Header().Get("Content-Type"))
			if tt.addr == "/api/user/login" {
				if result.StatusCode() == http.StatusOK {
					s := result.Cookies()
					cookie = *s[0]
				}
			}
		})
	}
	return cookie
}

func PutOrderTest(t *testing.T, app *App, cookie http.Cookie) {

	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name   string
		addr   string
		number string
		want   want
	}{
		{
			name:   "order upload ok",
			addr:   "/api/user/orders",
			number: "12345678903",
			want: want{
				statusCode:  http.StatusAccepted,
				contentType: "application/json",
			},
		},
		{
			name:   "order upload conflict",
			addr:   "/api/user/orders",
			number: "12345678903",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "order luhn failure",
			addr:   "/api/user/orders",
			number: "12345678902",
			want: want{
				statusCode:  http.StatusUnprocessableEntity,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := resty.New().R().SetHeader("Content-Type", "application/json").
				SetBody(tt.number).SetCookie(&cookie)

			result, err := request.Post("http://" + app.config.ServerAddress + tt.addr)
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode())
			assert.Equal(t, tt.want.contentType, result.Header().Get("Content-Type"))
		})
	}
}

func GetOrdersTest(t *testing.T, app *App, cookie http.Cookie) {

	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		addr string
		resp []service.Order
		want want
	}{
		{
			name: "get orders ok",
			addr: "/api/user/orders",
			resp: []service.Order{
				{
					Number: "12345678903",
					Status: storage.NEW,
				},
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var listOrders []service.Order
			request := resty.New().R().SetResult(&listOrders).
				SetHeader("Content-Type", "application/json").SetCookie(&cookie)

			result, err := request.Get("http://" + app.config.ServerAddress + tt.addr)
			require.NoError(t, err)

			tt.resp[0].UploadedAt = listOrders[0].UploadedAt
			assert.Equal(t, tt.resp, listOrders)
			assert.Equal(t, tt.want.statusCode, result.StatusCode())
			assert.Equal(t, tt.want.contentType, result.Header().Get("Content-Type"))
		})
	}
}

func GetBalanceTest(t *testing.T, app *App, cookie http.Cookie) {

	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name   string
		addr   string
		cookie bool
		resp   service.Balance
		want   want
	}{
		{
			name:   "get balance ok",
			addr:   "/api/user/balance",
			cookie: true,
			resp: service.Balance{
				Current:   0,
				Withdrawn: 0,
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "get balance unauth",
			addr:   "/api/user/balance",
			cookie: false,
			resp:   service.Balance{},
			want: want{
				statusCode:  http.StatusUnauthorized,
				contentType: "text/html; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var balance service.Balance
			request := resty.New().R().SetResult(&balance).SetHeader("Content-Type", "application/json")
			if tt.cookie {
				request.SetCookie(&cookie)
			}
			result, err := request.Get("http://" + app.config.ServerAddress + tt.addr)
			require.NoError(t, err)

			assert.Equal(t, tt.resp, balance)
			assert.Equal(t, tt.want.statusCode, result.StatusCode())
			assert.Equal(t, tt.want.contentType, result.Header().Get("Content-Type"))
		})
	}
}

func WithdrawTest(t *testing.T, app *App, cookie http.Cookie) {

	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name       string
		addr       string
		withdrawal service.Withdrawal
		want       want
	}{
		{
			name: "withdraw not enough",
			addr: "/api/user/balance/withdraw",
			withdrawal: service.Withdrawal{
				OrderID: "2377225624",
				Amount:  100,
			},
			want: want{
				statusCode:  http.StatusPaymentRequired,
				contentType: "",
			},
		},
		{
			name: "withdraw ok",
			addr: "/api/user/balance/withdraw",
			withdrawal: service.Withdrawal{
				OrderID: "2377225624",
				Amount:  0,
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := resty.New().R().SetBody(tt.withdrawal).
				SetHeader("Content-Type", "application/json").SetCookie(&cookie)

			result, err := request.Post("http://" + app.config.ServerAddress + tt.addr)
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode())
			assert.Equal(t, tt.want.contentType, result.Header().Get("Content-Type"))
		})
	}
}

func GetWithdrawalsTest(t *testing.T, app *App, cookie http.Cookie) {
	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		addr string
		resp []service.Withdrawal
		want want
	}{
		{
			name: "get orders ok",
			addr: "/api/user/withdrawals",
			resp: []service.Withdrawal{
				{
					OrderID: "2377225624",
					Amount:  0,
				},
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var listWithdrawals []service.Withdrawal
			request := resty.New().R().SetResult(&listWithdrawals).
				SetHeader("Content-Type", "application/json").SetCookie(&cookie)

			result, err := request.Get("http://" + app.config.ServerAddress + tt.addr)
			require.NoError(t, err)

			tt.resp[0].ProcessedAt = listWithdrawals[0].ProcessedAt
			assert.Equal(t, tt.resp, listWithdrawals)
			assert.Equal(t, tt.want.statusCode, result.StatusCode())
			assert.Equal(t, tt.want.contentType, result.Header().Get("Content-Type"))
		})
	}
}
