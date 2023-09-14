package routes

import (
	"chat-app-back/src/config"
	"chat-app-back/src/middlewares"
	"chat-app-back/src/models"
	"chat-app-back/src/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChangeUsername struct {
	Username string `json:"username" binding:"required"`
}

type ChangeStatus struct {
	CustomStatus string `json:"custom_status" binding:"required"`
}

type UserProfileResponse struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	CustomStatus   *string `json:"custom_status"`
	ProfilePicture *string `json:"profile_picture"`
}

func HandleChangeUsername(c *gin.Context) {
	db := config.MongoClient()

	var changeUsername ChangeUsername

	// Validate json structure
	err := c.BindJSON(&changeUsername)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}
	err = validate.Struct(changeUsername)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	// Get the sender's uid and attempt to change the username.
	uid := util.GetUid(c)
	objectID, err := primitive.ObjectIDFromHex(uid)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid user ID"})
		return
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"username": changeUsername.Username}}

	_, err = db.Database("Chat-App").Collection("users").UpdateOne(c, filter, update)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to change username"})
		return
	}

	c.JSON(200, gin.H{"status": "success", "message": "Username changed successfully"})
}

func HandleChangeStatus(c *gin.Context) {
	db := config.MongoClient()

	var changeStatus ChangeStatus

	// Validate json structure
	err := c.BindJSON(&changeStatus)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}
	err = validate.Struct(changeStatus)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	// Get the sender's uid and attempt to change the status.
	uid := util.GetUid(c)
	objectID, err := primitive.ObjectIDFromHex(uid)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid user ID"})
		return
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"custom_status": changeStatus.CustomStatus}}
	_, err = db.Database("Chat-App").Collection("users").UpdateOne(c, filter, update)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to change status"})
		return
	}
	c.JSON(200, gin.H{"status": "success", "message": "Status changed successfully"})
}

func HandleGetUserProfile(c *gin.Context) {
	db := config.MongoClient()

	// Get the user ID from the route parameter
	uid := c.Param("user_id")
	objectID, err := primitive.ObjectIDFromHex(uid)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid user ID"})
		return
	}

	// Find the user in the database
	var user models.User
	filter := bson.M{"_id": objectID}
	err = db.Database("Chat-App").Collection("users").FindOne(c, filter).Decode(&user)
	if err != nil {
		c.JSON(404, gin.H{"status": "error", "message": "User not found"})
		return
	}

	// Create a UserProfileResponse object
	userProfile := UserProfileResponse{
		ID:             user.ID.Hex(),
		Username:       user.Username,
		CustomStatus:   user.CustomStatus,
		ProfilePicture: user.ProfilePicture,
	}

	// Return the user profile as a JSON response
	c.JSON(200, gin.H{"status": "success", "user_profile": userProfile})
}
func ProfileRoutes(route *gin.RouterGroup) {
	profileGroup := route.Group("/")
	{
		profileGroup.POST("edit_username", middlewares.AuthenticateAccessToken(), HandleChangeUsername)
		profileGroup.POST("change_status", middlewares.AuthenticateAccessToken(), HandleChangeStatus)
		profileGroup.GET("user_profile/:user_id", HandleGetUserProfile)
	}
}
