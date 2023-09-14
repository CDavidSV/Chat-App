package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Channel struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty"`
	Name                *string            `bson:"name"`
	Type                string             `bson:"type"`
	Members             []string           `bson:"members"`
	CreatedAt           primitive.DateTime `bson:"created_at"`
	LastMessageID       primitive.DateTime `bson:"last_message_id"`
	OwnerID             *string            `bson:"owner_id"`
	GroupProfilePicture *string            `bson:"group_profile_picture"`
}
