package util

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func GetUid(c *gin.Context) string {
	claims, exists := c.Get("claims")
	if !exists {
		return ""
	}

	// Type assertion to *jwt.MapClaims
	claimsMap, ok := claims.(*jwt.MapClaims)
	if !ok || claimsMap == nil {
		return ""
	}

	// Dereference and type assert the uid
	uid, uidOk := (*claimsMap)["uid"].(string)
	if !uidOk || uid == "" {
		return ""
	}

	return uid
}
