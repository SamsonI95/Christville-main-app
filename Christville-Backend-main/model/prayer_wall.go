package model

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Prayer struct{
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id" json:"userId"`
	Prayer 		string				`bson:"prayer" json:"prayer"`
	CreatedAt	time.Time          `bson:"created_at" json:"createdAt"`
}