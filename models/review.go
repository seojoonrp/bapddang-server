// models/review.go

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewedFoodItem struct {
	FoodID primitive.ObjectID `bson:"foodId" json:"foodId"`
	FoodType string `bson:"foodType" json:"foodType"`
}

type Review struct {
	ID primitive.ObjectID `bson:"_id, omitempty" json:"id"`
	UserID primitive.ObjectID `bson:"userId" json:"userId" binding:"required"`
	
	Name string `bson:"name" json:"name" binding:"required"`
	Foods []ReviewedFoodItem `bson:"foods" json:"foods" binding:"required"`
	Speed string `bson:"speed" json:"speed" binding:"required"`
	MealTime string `bson:"mealTime" json:"mealTime" binding:"required"`

	Tags []string `bson:"tags,omitempty" json:"tags"`
	ImageURL string `bson:"imageUrl,omitempty" json:"imageUrl"`
	Comment string `bson:"comment,omitempty" json:"comment"`
	Rating int `bson:"rating,omitempty" json:"rating"`

	Day int `bson:"day" json:"day" binding:"required"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

type CreateReviewInput struct {
	Name string `json:"name" binding:"required"`
	Foods []ReviewedFoodItem `json:"foods" binding:"required"`
	Speed string `json:"speed" binding:"required"`
	MealTime string `json:"mealTime" binding:"required"`
	Tags []string `json:"tags"`
	ImageURL string `json:"imageUrl"`
	Comment string `json:"comment"`
	Rating *int `json:"rating"`
}