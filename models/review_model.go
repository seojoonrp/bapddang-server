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
	UserID primitive.ObjectID `bson:"userId" json:"userId"`
	
	Name string `bson:"name" json:"name"`
	Foods []ReviewedFoodItem `bson:"foods" json:"foods"`
	Speed string `bson:"speed" json:"speed"`
	MealTime string `bson:"mealTime" json:"mealTime"`

	Tags []string `bson:"tags" json:"tags"`
	ImageURL string `bson:"imageUrl" json:"imageUrl"`
	Comment string `bson:"comment" json:"comment"`
	Rating int `bson:"rating" json:"rating"`

	Day int `bson:"day" json:"day"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

type ReviewInput struct {
	Name string `json:"name" binding:"required"`
	Foods []ReviewedFoodItem `json:"foods" binding:"required"`
	Speed string `json:"speed" binding:"required"`
	MealTime string `json:"mealTime" binding:"required"`
	Tags []string `json:"tags"`
	ImageURL string `json:"imageUrl"`
	Comment string `json:"comment"`
	Rating int `json:"rating"`
}