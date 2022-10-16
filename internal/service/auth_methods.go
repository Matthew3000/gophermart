package service

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gophermart/internal/storage"
	"net/http"
	"regexp"
	"time"
)

type JWTToken struct {
	Login string `json:"login"`
	Auth  bool   `json:"authorized"`
	jwt.StandardClaims
}

func GenerateJWT(login string) (string, error) {
	var mySigningKey = []byte(secretKey)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["login"] = login
	claims["exp"] = time.Now().Add(time.Minute * 720).Unix()

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", fmt.Errorf("jwt generate: %s", err)
	}

	return tokenString, nil
}

func IsAuthorized(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//if r.Header["Token"] == nil {
		//	log.Printf("no token found")
		//	http.Error(w, "no token found", http.StatusBadRequest)
		//	return
		//}
		//
		//var mySigningKey = []byte(secretKey)
		//
		//token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
		//	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		//		return nil, fmt.Errorf("parsing token error")
		//	}
		//	return mySigningKey, nil
		//})
		//
		//if err != nil {
		//	log.Printf("token error: %s", err)
		//	http.Error(w, fmt.Sprintf("token error: %s", err), http.StatusUnauthorized)
		//	return
		//}
		//
		//if claims, ok := token.Claims.(*JWTToken); ok && token.Valid {
		//	log.Printf(claims.Login)
		//	r.Header.Set("login", claims.Login)
		//	handler.ServeHTTP(w, r)
		//}
		//http.Redirect(w, r, "/login", http.StatusUnauthorized)
		session, _ := storage.CookieStorage.Get(r, "session.id")
		authenticated := session.Values["authenticated"]
		if authenticated != nil && authenticated != false {
			handler.ServeHTTP(w, r)
		}
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
	}
}

func GeneratePasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CheckLoginAndPassword(user User) error {
	if user.Login == "" {
		return fmt.Errorf("empty login")
	}
	if user.Password == "" {
		return fmt.Errorf("empty password")
	}

	loginRegex := regexp.MustCompile(`^[\s\S]{6,}$`)
	if !loginRegex.MatchString(user.Login) {
		return fmt.Errorf("login should be at least 6 characters")
	}
	passwordRegex := regexp.MustCompile(`^[\s\S]{8,}$`)
	if !passwordRegex.MatchString(user.Password) {
		return fmt.Errorf("password should be at least 8 characters")
	}
	return nil
}
