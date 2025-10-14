package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	UserName string `bson:"userName"`
	Email string `bson:"email"`
	Password string `bson:"password"`
	CreatedAt time.Time `bson:"createdAt"`
	Day int `bsong:"day"`
	LikedFoodIDs []primitive.ObjectID `bson:"likedFoodIDs"`
}