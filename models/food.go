package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StandardFood struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string `bson:"name" json:"name" binding:"required"`
	ImageURL string `bson:"imageURL" json:"imageUrl" binding:"required"`

	Speed string `bson:"speed" json:"speed" binding:"required"`
	Type string `bson:"type" json:"type" binding:"required"`
	Categories []string `bson:"categories" json:"categories"`

	LikeCount int `bson:"likeCount" json:"likeCount"`
	ReviewCount int `bson:"reviewCount" json:"reviewCount"`
	AverageRating float64 `bson:"averageRating" json:"averageRating"`

	TrendScore int `bson:"trendScore" json:"trendScore"`
}

type CustomFood struct {
	ID primitive.ObjectID `bson:"_id, omitempty" json:"id"`
	Name string `bson:"name" json:"name" binding:"required"`
	UsingUserIDs []primitive.ObjectID `bson:"usingUserIDs" json:"usingUserIDs" binding:"required"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}