package routes

import (
	config "chat-app-back/src/config"
	models "chat-app-back/src/models"
	"chat-app-back/src/util"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GoogleToken struct {
	Token string `json:"account_token" validate:"required"`
}

type AccountCreation struct {
	Username string `json:"username" validate:"required"`
	Token    string `json:"account_token" validate:"required"`
}

type RefreshTokenBody struct {
	Token string `json:"refresh_token" validate:"required"`
}

var validate = validator.New()

// Handles authenticating valid refresh tokens
func authenticateRefreshToken(refreshToken string) (string, error) {
	// Get token secret from env file
	tokenSecret := os.Getenv("REFRESH_TOKEN_KEY")

	// Initialize instance of Claims
	claims := &jwt.MapClaims{}

	// Attempt to parse the token
	parsedToken, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})

	// Check if token was parsed
	if err != nil || !parsedToken.Valid {
		return "", err
	}

	return (*claims)["uid"].(string), nil
}

// Handles generating access and refresh tokens
func generateTokens(uid string) (string, string) {
	var expirationTime int64 = 60 * 60 * 24 // 24 hours in seconds
	accessToken, err := util.GenerateToken(uid, expirationTime, false)
	if err != nil {
		return "", ""
	}

	// Add 7 days to expiration time for refresh tokens
	refreshToken, err := util.GenerateToken(uid, expirationTime*7, true)
	if err != nil {
		return "", ""
	}

	return accessToken, refreshToken
}

func insertNewRefreshToken(userId string, refreshToken string, expirationTime int64) bool {
	db := config.MongoClient()

	// Save refresh token to database
	c, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := db.Database("Chat-App").Collection("refresh_tokens").InsertOne(c, models.RefreshToken{
		ID:           primitive.NewObjectID(),
		UserID:       userId,
		RefreshToken: refreshToken,
		CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		ExpiresAt:    primitive.NewDateTimeFromTime(time.Now().Add(time.Second * time.Duration(expirationTime)))})

	return err == nil
}

func updateRefreshToken(oldRefreshToken string, newRefreshToken string, userId string, expirationTime int64) bool {
	db := config.MongoClient()

	// Save refresh token to database
	c, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := db.Database("Chat-App").Collection("refresh_tokens").FindOneAndUpdate(c,
		bson.M{
			"user_id":       userId,
			"refresh_token": oldRefreshToken},
		bson.D{{
			Key: "$set",
			Value: bson.M{
				"refresh_token": newRefreshToken,
				"created_at":    primitive.NewDateTimeFromTime(time.Now()),
				"expires_at":    primitive.NewDateTimeFromTime(time.Now().Add(time.Second * time.Duration(expirationTime)))}}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Err()

	fmt.Println(err)
	return err == nil
}

func HandleLogin(c *gin.Context) {
	admin := config.InitializeApp()
	db := config.MongoClient()
	var googleAccountToken GoogleToken

	err := c.BindJSON(&googleAccountToken)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}
	err = validate.Struct(googleAccountToken)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	// Get auth client
	client, err := admin.Auth(c)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Error getting Auth client"})
		return
	}

	// Verify google account token
	token, err := client.VerifyIDToken(c, googleAccountToken.Token)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Error verifying account token"})
		return
	}

	// Get user data
	var result models.User
	err = db.Database("Chat-App").Collection("users").FindOneAndUpdate(c, bson.D{{Key: "email", Value: token.Claims["email"]}}, bson.M{"$set": bson.M{"profile_picture": token.Claims["picture"]}}).Decode(&result)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"status": "error", "message": "User does not exist"})
		return
	}

	// Generate access and refresh tokens.
	accessToken, refreshToken := generateTokens(result.ID.Hex())
	if accessToken == "" || refreshToken == "" {
		c.JSON(500, gin.H{"status": "error", "message": "Error generating tokens"})
		return
	}
	inserted := insertNewRefreshToken(result.ID.Hex(), refreshToken, 60*60*24*7)
	if !inserted {
		c.JSON(500, gin.H{"status": "error", "message": "Error updating refresh token"})
		return
	}

	c.JSON(200, gin.H{"status": "success", "message": "User logged in successfully", "access_token": accessToken, "refresh_token": refreshToken})

	// Set offline status after 15 minutes
	util.SetOfflineAfterDuration(result.ID.Hex(), 15*time.Minute, c)
}

func HandleRegister(c *gin.Context) {
	admin := config.InitializeApp()
	db := config.MongoClient()
	var accountCreation AccountCreation

	err := c.BindJSON(&accountCreation)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}
	err = validate.Struct(accountCreation)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	// Get auth client
	client, err := admin.Auth(c)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Error getting Auth client"})
		return
	}

	// Verify google account token
	token, err := client.VerifyIDToken(c, accountCreation.Token)

	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Error verifying account token"})
		return
	}

	var result models.User
	err = db.Database("Chat-App").Collection("users").FindOne(c, bson.D{{Key: "email", Value: token.Claims["email"]}}).Decode(&result)

	if err == nil {
		// User already exists
		c.JSON(500, gin.H{"status": "error", "message": "User already exists"})
		return
	}

	// Create user
	profilePicture := token.Claims["picture"].(string)
	newUser, err := db.Database("Chat-App").Collection("users").InsertOne(c, models.User{
		ID:             primitive.NewObjectID(),
		FirebaseID:     token.UID,
		Email:          token.Claims["email"].(string),
		Username:       accountCreation.Username,
		ProfilePicture: &profilePicture,
		Status:         "online",
		CustomStatus:   nil})
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Error creating user"})
		return
	}

	// Generate access and refresh tokens
	accessToken, refreshToken := generateTokens(newUser.InsertedID.(primitive.ObjectID).Hex())
	if accessToken == "" || refreshToken == "" {
		c.JSON(500, gin.H{"status": "error", "message": "Error generating tokens"})
		return
	}

	inserted := insertNewRefreshToken(newUser.InsertedID.(primitive.ObjectID).Hex(), refreshToken, 60*60*24*7)
	if !inserted {
		c.JSON(500, gin.H{"status": "error", "message": "Error updating refresh token"})
		return
	}

	c.JSON(200, gin.H{"status": "success", "message": "User created successfully", "access_token": accessToken, "refresh_token": refreshToken})

	// Set offline status after 15 minutes
	util.SetOfflineAfterDuration(newUser.InsertedID.(primitive.ObjectID).Hex(), 15*time.Minute, c)
}

func HandleRevokeToken(c *gin.Context) {
	var refreshToken RefreshTokenBody
	db := config.MongoClient()

	err := c.BindJSON(&refreshToken)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}
	if refreshToken.Token == "" {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	// Get uid from refresh token
	uid, err := authenticateRefreshToken(refreshToken.Token)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "invalid refresh token"})
		return
	}

	// Revoke refresh token
	revokedToken, err := db.Database("Chat-App").Collection("refresh_tokens").DeleteOne(c, bson.M{"user_id": uid, "refresh_token": refreshToken.Token})
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Error revoking refresh token"})
		return
	}
	if revokedToken.DeletedCount == 0 {
		c.JSON(500, gin.H{"status": "error", "message": "Invalid refresh token"})
		return
	}

	c.JSON(200, gin.H{"status": "success", "message": "Token revoked successfully"})

	// Set offline status
	util.SetOfflineAfterDuration(uid, 30*time.Second, c)
}

func HandleRefreshToken(c *gin.Context) {
	var refreshToken RefreshTokenBody
	db := config.MongoClient()

	err := c.BindJSON(&refreshToken)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}
	if refreshToken.Token == "" {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	// Validate the token
	uid, err := authenticateRefreshToken(refreshToken.Token)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "invalid refresh token"})
		return
	}

	// Check if the refresh token exists in the database
	var result models.RefreshToken
	err = db.Database("Chat-App").Collection("refresh_tokens").FindOne(c, bson.M{"user_id": uid, "refresh_token": refreshToken.Token}).Decode(&result)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "invalid refresh token"})
		return
	}

	// Delete the refresh token from the database
	db.Database("Chat-App").Collection("refresh_tokens").DeleteOne(c, bson.M{"user_id": uid, "refresh_token": refreshToken.Token})

	// Generate new tokens
	accessToken, newRefreshToken := generateTokens(uid)
	if accessToken == "" || newRefreshToken == "" {
		c.JSON(500, gin.H{"status": "error", "message": "Error generating tokens"})
		return
	}

	updated := updateRefreshToken(refreshToken.Token, newRefreshToken, uid, 60*60*24*7)
	if !updated {
		c.JSON(500, gin.H{"status": "error", "message": "Error updating refresh token"})
		return
	}

	c.JSON(200, gin.H{"status": "success", "message": "Token refreshed successfully", "access_token": accessToken, "refresh_token": newRefreshToken})

	// Set offline status after 15 minutes
	util.SetOfflineAfterDuration(uid, 15*time.Minute, c)
}

func AuthenticationRoutes(route *gin.RouterGroup) {
	authGroup := route.Group("/")
	{
		authGroup.POST("login", HandleLogin)
		authGroup.POST("register", HandleRegister)
		authGroup.POST("revoke", HandleRevokeToken)
		authGroup.POST("refresh_token", HandleRefreshToken)
	}
}
