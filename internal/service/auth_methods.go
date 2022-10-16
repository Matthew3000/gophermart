package service

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"regexp"
)

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
