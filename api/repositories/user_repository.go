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
	FindByUsername(username string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Save(user *models.User) error
	AddLikedFood(userID, foodID primitive.ObjectID) (bool, error)
	RemoveLikedFood(userID, foodID primitive.ObjectID) (bool, error)
	GetLikedFoodIDs(userID primitive.ObjectID) ([]primitive.ObjectID, error)
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(coll *mongo.Collection) UserRepository {
	return &userRepository{collection: coll}
}

func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
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

func (r *userRepository) AddLikedFood(userID, foodID primitive.ObjectID) (bool, error) {
	filter := bson.M{"_id": userID}
	update := bson.M{"$addToSet": bson.M{"likedFoodIDs": foodID}}

	result, err := r.collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return false, err
	}
	return result.ModifiedCount > 0, nil
}

func (r *userRepository) RemoveLikedFood(userID, foodID primitive.ObjectID) (bool, error) {
	filter := bson.M{"_id": userID}
	update := bson.M{"$pull": bson.M{"likedFoodIDs": foodID}}

	result, err := r.collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return false, err
	}
	return result.ModifiedCount > 0, nil
}

func (r *userRepository) GetLikedFoodIDs(userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	var user models.User
	err := r.collection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user.LikedFoodIDs, nil
}
