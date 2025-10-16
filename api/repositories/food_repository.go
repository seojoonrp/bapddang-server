// api/repositories/food_repository.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FoodRepository interface {
	FindStandardFoodByID(id primitive.ObjectID) (*models.StandardFood, error)
	SaveStandardFood(food *models.StandardFood) error
	FindCustomFoodByName(name string) (*models.CustomFood, error)
	SaveCustomFood(food *models.CustomFood) error
	AddUserToCustomFood(foodID, userID primitive.ObjectID) error
}

type foodRepository struct {
	standardFoodCollection *mongo.Collection
	customFoodCollection *mongo.Collection
}

func NewFoodRepository(standardColl *mongo.Collection, customColl *mongo.Collection) FoodRepository {
	return &foodRepository{
		standardFoodCollection: standardColl,
		customFoodCollection: customColl,
	}
}

func (r *foodRepository) FindStandardFoodByID(id primitive.ObjectID) (*models.StandardFood, error) {
	var food models.StandardFood
	err := r.standardFoodCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&food)
	if err != nil {
		return nil, err
	}
	return &food, nil
}

func (r *foodRepository) SaveStandardFood(food *models.StandardFood) error {
	_, err := r.standardFoodCollection.InsertOne(context.TODO(), food)
	return err
}

func (r *foodRepository) FindCustomFoodByName(name string) (*models.CustomFood, error) {
	var food models.CustomFood
	err := r.customFoodCollection.FindOne(context.TODO(), bson.M{"name": name}).Decode(&food)
	if err != nil {
		return nil, err
	}
	return &food, nil
}

func (r *foodRepository) SaveCustomFood(food *models.CustomFood) error {
	_, err := r.customFoodCollection.InsertOne(context.TODO(), food)
	return err
}

func (r *foodRepository) AddUserToCustomFood(foodID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": foodID}
	update := bson.M{"$addToSet": bson.M{"usingUserIDs": userID}}
	_, err := r.customFoodCollection.UpdateOne(context.TODO(), filter, update)
	return err
}