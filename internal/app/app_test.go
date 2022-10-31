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
	"log"
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
	log.Printf(cookie.Value)
	app.userStorage.DeleteAll()
}

func RegisterTest(t *testing.T, app *App) http.Cookie {
	var cookie http.Cookie
	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name    string
		addr    string
		method  string
		handler http.HandlerFunc
		user    service.User
		want    want
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
		name    string
		addr    string
		handler http.HandlerFunc
		number  string
		want    want
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
			request := resty.New().R().SetHeader("Content-Type", "application/json").SetBody(tt.number).SetCookie(&cookie)

			result, err := request.Post("http://" + app.config.ServerAddress + tt.addr)
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode())
			assert.Equal(t, tt.want.contentType, result.Header().Get("Content-Type"))
		})
	}
}
