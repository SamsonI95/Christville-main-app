package model

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type BibleVerse struct {
	Reference string `json:"reference"`
    Text string `json:"text"`
}

type CachedVerse struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Reference string             `bson:"reference"` 
    Text      string             `bson:"text"`    
    Timestamp time.Time          `bson:"timestamp"`
}