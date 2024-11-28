package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Referral struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ReferrerID     primitive.ObjectID `bson:"referrer_id" json:"referrerId"`
	ReferredID     primitive.ObjectID `bson:"referred_id" json:"referredId"`
	DirectReferrer bool               `bson:"direct_referrer" json:"directReferrer"`
	CoinsEarned    int                `bson:"coins_earned" json:"coinsEarned"`
	LastRewardTime time.Time          `bson:"last_reward_time" json:"lastRewardTime"`
	CreatedAt      time.Time          `bson:"created_at" json:"createdAt"`
}

func NewReferral(referrerID, referredID primitive.ObjectID) *Referral {
	return &Referral{
		ReferrerID:     referrerID,
		ReferredID:     referredID,
		CoinsEarned:    0,
		LastRewardTime: time.Time{},
		CreatedAt:      time.Now(),
	}
}
