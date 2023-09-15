package util

import (
	"chat-app-back/src/config"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userTimers = make(map[string]*time.Timer)

func SetOfflineAfterDuration(uid string, d time.Duration, c *gin.Context) {
	db := config.MongoClient()
	if timer, exists := userTimers[uid]; exists {
		timer.Stop()
	}

	// Update status to online
	objectID, err := primitive.ObjectIDFromHex(uid)
	if err != nil {
		log.Fatalln(err)
		return
	}
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"status": "online"}}
	_, err = db.Database("Chat-App").Collection("users").UpdateOne(c, filter, update)
	if err != nil {
		log.Fatalln(err)
		return
	}

	timer := time.AfterFunc(d, func() {
		if _, exists := userTimers[uid]; exists {
			// Change status on db

			filter := bson.M{"_id": objectID}
			update := bson.M{"$set": bson.M{"status": "offline"}}
			db.Database("Chat-App").Collection("users").UpdateOne(c, filter, update)
		}
	})
	userTimers[uid] = timer
}
