package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthenticateAccessToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Authorization header required"})
			c.Abort()
			return
		}

		// Split the header to remove Bearer
		tokenString := strings.Split(authHeader, " ")[1]

		// Get token secret from env file
		tokenSecret := os.Getenv("ACCESS_TOKEN_KEY")

		// Initialize instance of Claims
		claims := &jwt.MapClaims{}

		// Attempt to parse the token
		parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		})

		// Check if token was parsed
		if err != nil || !parsedToken.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid token"})
			c.Abort()
			return
		}

		// Get uid from claims
		uid := (*claims)["uid"].(string)
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid token"})
			c.Abort()
			return
		}

		// Pass the claims to the next handler
		c.Set("claims", claims)
		c.Next() // Call next handler
	}
}
