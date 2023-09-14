package routes

import (
	"chat-app-back/src/config"
	"chat-app-back/src/middlewares"
	"chat-app-back/src/models"
	"chat-app-back/src/util"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	// Check if channel exists
	id, err := primitive.ObjectIDFromHex(messageContent.ChannelId)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid channel id"})
		return
	}

	var channel models.Channel
	err = db.Database("Chat-App").Collection("channels").FindOne(c, bson.D{{Key: "_id", Value: id}}).Decode(&channel)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Channel does not exist"})
		return
	}

	message := models.Message{
		ID:        primitive.NewObjectID(),
		SenderID:  uid,
		ChannelID: messageContent.ChannelId,
		Content:   messageContent.Content,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now())}
	_, err = db.Database("Chat-App").Collection("messages").InsertOne(c, message)

	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to send message"})
		return
	}

	c.JSON(200, gin.H{"status": "success", "message": "Message sent successfully", "at": message.CreatedAt})
}

func HandleCreateChannel(c *gin.Context) {
	db := config.MongoClient()

	var channelCreate ChannelCreate

	// Validate json structure
	err := c.BindJSON(&channelCreate)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}
	err = validate.Struct(channelCreate)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	// Validate channel type
	if strings.ToLower(channelCreate.Type) != "dm" {
		c.JSON(400, gin.H{"status": "error", "message": "Unsupported channel type"})
		return
	}

	// Get the sender's uid.
	uid := util.GetUid(c)

	// Check if a channel already exists between the sender and the recepient
	var channel models.Channel
	err = db.Database("Chat-App").Collection("channels").FindOne(c, bson.D{{Key: "$and", Value: []bson.M{
		{"type": "dm"},
		{"members": channelCreate.RecepientID},
		{"members": uid}}}}).Decode(&channel)
	if err == nil {
		c.JSON(200, gin.H{"status": "error", "message": "Channel already exists", "channel_data": channel})
		return
	}

	// Check that the recepient is in the friendlist of the sender
	cursor, err := db.Database("Chat-App").Collection("friends").Find(c, bson.D{{Key: "$or", Value: []bson.M{
		{"user1_id": uid},
		{"user2_id": uid}}}})
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to create channel"})
		return
	}

	var createChannel bool = false
	for cursor.Next(c) {
		var friend models.Friend
		err := cursor.Decode(&friend)
		if err != nil {
			log.Fatal(err)
		}

		if friend.User1ID == channelCreate.RecepientID || friend.User2ID == channelCreate.RecepientID {
			createChannel = true
			break
		}
	}

	if !createChannel {
		c.JSON(400, gin.H{"status": "error", "message": "Recepient is not in your friendlist"})
		return
	}

	// Create channel
	channel = models.Channel{
		ID:   primitive.NewObjectID(),
		Type: "dm",
		Members: []string{
			uid,
			channelCreate.RecepientID},
		Name:                nil,
		OwnerID:             nil,
		GroupProfilePicture: nil,
		CreatedAt:           primitive.NewDateTimeFromTime(time.Now())}

	_, err = db.Database("Chat-App").Collection("channels").InsertOne(c, channel)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"status": "error", "message": "Failed to create channel"})
		return
	}
	c.JSON(200, gin.H{"status": "success", "message": "Channel created successfully", "channel_data": channel})
}

func HandleGetMessages(c *gin.Context) {
	db := config.MongoClient()

	channel_id := c.Query("channel_id")
	if channel_id == "" {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request"})
		return
	}

	id, err := primitive.ObjectIDFromHex(channel_id)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid channel id"})
		return
	}

	// Get uid
	uid := util.GetUid(c)

	var channel models.Channel
	err = db.Database("Chat-App").Collection("channels").FindOne(c, bson.D{{Key: "$and", Value: []bson.M{{"_id": id}, {"members": uid}}}}).Decode(&channel)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Channel does not exist"})
		return
	}

	// Get all messages in the channel
	cursor, err := db.Database("Chat-App").Collection("messages").Find(c, bson.D{{Key: "channel_id", Value: channel_id}})
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

	c.JSON(200, gin.H{"status": "success", "channel_id": channel.ID.Hex(), "messages": messages})
}

func HandleGetChannels(c *gin.Context) {
	// db := config.MongoClient()

	// channel_id := c.Query("channel_id")
	// if channel_id == "" {
	// 	c.JSON(400, gin.H{"status": "error", "message": "Invalid request"})
	// 	return
	// }

	// id, err := primitive.ObjectIDFromHex(channel_id)
	// if err != nil {
	// 	c.JSON(400, gin.H{"status": "error", "message": "Invalid channel id"})
	// 	return
	// }

	// // Get uid
	// uid := util.GetUid(c)
}

func MessageRoutes(route *gin.RouterGroup) {
	messagesGroup := route.Group("/")
	{
		messagesGroup.POST("send_message", middlewares.AuthenticateAccessToken(), HandleSendMessage)
		messagesGroup.POST("create_channel", middlewares.AuthenticateAccessToken(), HandleCreateChannel)
		messagesGroup.GET("get_messages", middlewares.AuthenticateAccessToken(), HandleGetMessages)
		messagesGroup.GET("get_channels", middlewares.AuthenticateAccessToken(), HandleGetChannels)
	}
}
