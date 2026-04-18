package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var SecretKey []byte

func GenerateToken(userID uint) (string, error) {
	key := os.Getenv("JWT_SECRET")

	if key == "" {
		key = "my-default-secret"
	}

	SecretKey = []byte(key)

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(SecretKey)
}
