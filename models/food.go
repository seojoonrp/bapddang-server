package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type StandardFood struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	Name string `bson:"name"`
	ImageURL string `bson:"imageURL"`

	Speed string `bson:"speed"`
	Category string `bson:"category"`
	Parents []string `bson:"parents"`

	LikeCount int `bson:"likeCount"`
	ReviewCount int `bson:"reviewCount"`
	AverageRating float64 `bson:"averageRating"`

	TrendScore int `bson:"trendScore"`
}