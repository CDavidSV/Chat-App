package routes

import (
	"chat-app-back/src/config"
	"chat-app-back/src/middlewares"
	"chat-app-back/src/models"
	"chat-app-back/src/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GetMessageContent struct {
	Content   string `json:"content"`
	ChannelId string `json:"channel_id"`
}

type ChannelCreate struct {
	Type        string `json:"type"`
	RecepientID string `json:"recepient_id"`
}

type MessageContent struct {
	ID        string             `json:"id"`
	SenderID  string             `json:"sender_id"`
	Me        bool               `json:"me"`
	CreatedAt primitive.DateTime `json:"created_at"`
	Content   string             `json:"content"`
}

var validate = validator.New()

func HandleSendMessage(c *gin.Context) {
	db := config.MongoClient()

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

	c.JSON(200, gin.H{"status": "success", "message": "Message sent successfully", "at": message.CreatedAt})
}

func HandleGetMessages(c *gin.Context) {
	db := config.MongoClient()

	// Get uid
	uid := util.GetUid(c)

	// Get all messages in the channel
	cursor, err := db.Database("Chat-App").Collection("messages").Find(c, bson.M{}, options.Find().SetLimit(100))
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to fetch messages"})
		return
	}

	var messages []MessageContent

	for cursor.Next(c) {
		var message models.Message
		cursor.Decode(&message)

		messageContent := MessageContent{
			ID:        message.ID.Hex(),
			SenderID:  message.SenderID,
			Me:        message.SenderID == uid,
			CreatedAt: message.CreatedAt,
			Content:   message.Content}

		messages = append(messages, messageContent)
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
