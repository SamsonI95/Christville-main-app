package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReferralEarnings struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ReferrerID    primitive.ObjectID `bson:"referrer_id" json:"referrerId"`
	ReferredID    primitive.ObjectID `bson:"referred_id" json:"referredId"`
	CoinsEarned   int                `bson:"coins_earned" json:"coinsEarned"`
	XPEarned      float64            `bson:"xp_earned" json:"xpEarned"`
	AirdropPoints int                `bson:"airdrop_points" json:"airdropPoints"`
	LastEarnedAt  time.Time          `bson:"last_earned_at" json:"lastEarnedAt"`
	CreatedAt     time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updatedAt"`
}
