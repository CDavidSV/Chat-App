package routes

import (
	config "chat-app-back/src/config"
	models "chat-app-back/src/models"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GoogleToken struct {
	Token string `json:"account_token"`
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

	// Get auth client
	client, err := admin.Auth(c)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Error getting Auth client"})
		fmt.Println(err)
		return
	}

	// Verify google account token
	token, err := client.VerifyIDToken(c, googleAccountToken.Token)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Error verifying account token"})
		return
	}

	// Get user data and save to database
	var result models.User
	err = db.Database("Chat-App").Collection("users").FindOne(c, bson.D{{Key: "email", Value: token.Claims["email"]}}).Decode(&result)

	if err == nil {
		c.JSON(200, gin.H{"status": "success", "message": "User loggen in"})
		return
	}

	// Create user
	db.Database("Chat-App").Collection("users").InsertOne(c, models.User{
		ID:         primitive.NewObjectID(),
		FirebaseID: token.UID,
		Email:      token.Claims["email"].(string),
		Username:   token.Claims["name"].(string),
		Status:     "offline", CustomStatus: "",
		ProfilePicture: token.Claims["picture"].(string)})
}

func HandleRegister(c *gin.Context) {}

func HandleRevokeToken(c *gin.Context) {}

func HandleRefreshToken(c *gin.Context) {}

func AuthenticationRoutes(route *gin.RouterGroup) {
	authGroup := route.Group("/")
	{
		authGroup.POST("login", HandleLogin)
		authGroup.POST("register", HandleRegister)
		authGroup.POST("revoke", HandleRevokeToken)
		authGroup.POST("refresh_token", HandleRefreshToken)
	}
}
