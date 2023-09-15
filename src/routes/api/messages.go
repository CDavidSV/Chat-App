package routes

import (
	"chat-app-back/src/config"
	"chat-app-back/src/middlewares"
	"chat-app-back/src/models"
	"chat-app-back/src/util"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GetMessageContent struct {
	Content   string `json:"content"`
	ChannelId string `json:"channel_id"`
}

type ChannelCreate struct {
	Type        string `json:"type"`
	RecepientID string `json:"recepient_id"`
}

type MessageUser struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	ProfilePicture string `json:"profile_picture"`
}

type MessageContent struct {
	ID        string             `json:"id"`
	SenderID  string             `json:"sender_id"`
	CreatedAt primitive.DateTime `json:"created_at"`
	Content   string             `json:"content"`
	User      MessageUser        `json:"user"`
}

var validate = validator.New()

func HandleSendMessage(c *gin.Context) {
	db := config.MongoClient()
	pusherClient := config.PusherInit()

	var messageContent GetMessageContent

	// Validate json structure
	err := c.BindJSON(&messageContent)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}
	err = validate.Struct(messageContent)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	// Get the sender's uid and attempt to send the message.
	uid := util.GetUid(c)

	objectID, err := primitive.ObjectIDFromHex(uid)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid user ID"})
		return
	}

	message := models.Message{
		ID:        primitive.NewObjectID(),
		SenderID:  uid,
		Content:   messageContent.Content,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now())}
	_, err = db.Database("Chat-App").Collection("messages").InsertOne(c, message)

	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to send message"})
		return
	}

	// Fetch User data
	var user models.User
	filter := bson.M{"_id": objectID}
	err = db.Database("Chat-App").Collection("users").FindOne(c, filter).Decode(&user)
	if err != nil {
		fmt.Println(err.Error())
	}

	c.JSON(200, gin.H{"status": "success", "message": "Message sent successfully", "at": message.CreatedAt})

	go func() {
		// Trigger pusher event
		data := map[string]any{
			"id":         message.ID.Hex(),
			"sender_id":  message.SenderID,
			"me":         false,
			"created_at": message.CreatedAt,
			"content":    message.Content,
			"user": map[string]string{
				"id":              uid,
				"username":        user.Username,
				"profile_picture": *user.ProfilePicture,
			},
		}
		err = pusherClient.Trigger("super-chat-channel", "main", data)
		if err != nil {
			fmt.Println(err.Error())
		}

		// Set offline status after 15 minutes
		util.SetOfflineAfterDuration(uid, 15*time.Minute, c)
	}()
}

func HandleGetMessages(c *gin.Context) {
	db := config.MongoClient()

	// Create pipeline to fetch user data for messages
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$sort", Value: bson.M{"created_at": -1}}},
		bson.D{{Key: "$limit", Value: 100}},
		{{Key: "$addFields", Value: bson.M{"sender_id_object": bson.M{"$toObjectId": "$sender_id"}}}},
		bson.D{{
			Key: "$lookup", Value: bson.M{
				"from":         "users",            // The other collection
				"localField":   "sender_id_object", // Name of the field in messages collection
				"foreignField": "_id",              // Name of the field in users collection
				"as":           "user",             // Output array field
			},
		}},
	}

	// Get all messages in the channel
	cursor, err := db.Database("Chat-App").Collection("messages").Aggregate(c, pipeline)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to fetch messages"})
		return
	}

	var messages []MessageContent

	for cursor.Next(c) {
		var raw bson.M
		if err := cursor.Decode(&raw); err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Failed to fetch messages"})
			return
		}

		messageContent := MessageContent{
			ID:        raw["_id"].(primitive.ObjectID).Hex(),
			SenderID:  raw["sender_id"].(string),
			CreatedAt: raw["created_at"].(primitive.DateTime),
			Content:   raw["content"].(string),
			User: MessageUser{
				ID:             raw["user"].(primitive.A)[0].(bson.M)["_id"].(primitive.ObjectID).Hex(),
				Username:       raw["user"].(primitive.A)[0].(bson.M)["username"].(string),
				ProfilePicture: raw["user"].(primitive.A)[0].(bson.M)["profile_picture"].(string),
			},
		}

		messages = append(messages, messageContent)
	}

	// Reverse the messages slice
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	c.JSON(200, gin.H{"status": "success", "messages": messages})
}

func MessageRoutes(route *gin.RouterGroup) {
	messagesGroup := route.Group("/")
	{
		messagesGroup.POST("send_message", middlewares.AuthenticateAccessToken(), HandleSendMessage)
		messagesGroup.GET("get_messages", middlewares.AuthenticateAccessToken(), HandleGetMessages)
	}
}
