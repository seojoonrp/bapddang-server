// models/user.go

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserName string `bson:"userName" json:"userName" binding:"required"`
	Email string `bson:"email" json:"email" binding:"required"`
	Password string `bson:"password" json:"-"`
	Day int `bson:"day" json:"day"`
	LikedFoodIDs []primitive.ObjectID `bson:"likedFoodIDs" json:"likedFoodIDs"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}