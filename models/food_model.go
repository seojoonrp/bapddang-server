// models/food.go

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
	RatedReviewCount int `bson:"ratedReviewCount" json:"ratedReviewCount"`
	TotalRating int `bson:"totalRating" json:"totalRating"`
	AverageRating float64 `bson:"-" json:"averageRating"`

	TrendScore int `bson:"trendScore" json:"trendScore"`
}

type NewStandardFoodInput struct {
	Name string `json:"name"`
	ImageURL string `json:"imageURL"`
	Speed string `json:"speed"`
	Type string `json:"type"`
	Categories []string `json:"categories"`
}

type CustomFood struct {
	ID primitive.ObjectID `bson:"_id, omitempty" json:"id"`
	Name string `bson:"name" json:"name" binding:"required"`
	UsingUserIDs []primitive.ObjectID `bson:"usingUserIDs" json:"usingUserIDs" binding:"required"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

type NewCustomFoodInput struct {
	Name string `json:"name"`
}

type ValidateFoodsInput struct {
	Names []string `json:"names" binding:"required,min=1"`
}

type ValidationResult struct {
	Status string `json:"status"`
	OriginalName string `json:"originalName"`
	Food any `json:"food,omitempty"`
	Suggestion any `json:"suggestion,omitempty"`
}