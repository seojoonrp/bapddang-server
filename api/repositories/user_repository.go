// api/repositories/user_repository.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	Save(user *models.User) error
	AddLikedFood(userID, foodID primitive.ObjectID) error
	RemoveLikedFood(userID, foodID primitive.ObjectID) error
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(coll *mongo.Collection) UserRepository {
	return &userRepository{collection: coll}
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Save(user *models.User) error {
	_, err := r.collection.InsertOne(context.TODO(), user)
	return err
}

func (r *userRepository) AddLikedFood(userID, foodID primitive.ObjectID) error {
	filter := bson.M{"_id": userID}
	update := bson.M{"$addToSet": bson.M{"likedFoodIDs": foodID}}

	_, err := r.collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *userRepository) RemoveLikedFood(userID, foodID primitive.ObjectID) error {
	filter := bson.M{"_id": userID}
	update := bson.M{"$pull": bson.M{"likedFoodIDs": foodID}}

	_, err := r.collection.UpdateOne(context.TODO(), filter, update)
	return err
}