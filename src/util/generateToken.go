package util

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func GenerateToken(uid string, expirationDelta int64, refresh bool) (string, error) {
	// Get toke secret from env file
	var tokenSecret string
	if refresh {
		tokenSecret = os.Getenv("ACCESS_TOKEN_KEY")
	} else {
		tokenSecret = os.Getenv("REFRESH_TOKEN_KEY")
	}

	// Generate the new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": uid,
		"exp": time.Now().Add(time.Second * time.Duration(expirationDelta)).Unix(),
	})

	// Sign the token
	tokenString, err := token.SignedString([]byte(tokenSecret))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
