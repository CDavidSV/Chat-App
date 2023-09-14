package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Friend struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	User1ID string             `bson:"user1_id"`
	User2ID string             `bson:"user2_id"`
	Since   primitive.DateTime `bson:"since"`
}
