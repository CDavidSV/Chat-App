package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type RefreshToken struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       string             `bson:"user_id,omitempty"`
	RefreshToken string             `bson:"refresh_token,omitempty"`
	CreatedAt    primitive.DateTime `bson:"created_at,omitempty"`
	ExpiresAt    primitive.DateTime `bson:"expires_at,omitempty"`
}
