package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type FriendRequest struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	SenderID  string             `bson:"sender_id"`
	InviteeID string             `bson:"invitee_id"`
}
