package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	FirebaseID     string             `bson:"firebase_id,omitempty"`
	Email          string             `bson:"email,omitempty"`
	Username       string             `bson:"username,omitempty"`
	CustomStatus   string             `bson:"custom_status,omitempty"`
	ProfilePicture string             `bson:"profile_picture,omitempty"`
}
