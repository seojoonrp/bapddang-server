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
	FindStandardFoodByName(name string) (*models.StandardFood, error)
	FindCustomFoodByName(name string) (*models.CustomFood, error)
	GetAllStandardFoods() ([]*models.StandardFood, error)
	GetAllCustomFoods() ([]*models.CustomFood, error)

	SaveCustomFood(food *models.CustomFood) error
	SaveStandardFood(food *models.StandardFood) error
	
	AddUserToCustomFood(foodID, userID primitive.ObjectID) error
	UpdateCreatedReviewStats(foodID []primitive.ObjectID, rating int) error
	UpdateModifiedReviewStats(foodID []primitive.ObjectID, oldRating, newRating int) error
	IncrementLikeCount(foodID primitive.ObjectID) error
	DecrementLikeCount(foodID primitive.ObjectID) error
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

func (r *foodRepository) FindStandardFoodByName(name string) (*models.StandardFood, error) {
	var food models.StandardFood
	err := r.standardFoodCollection.FindOne(context.TODO(), bson.M{"name": name}).Decode(&food)
	if err != nil {
		return nil, err
	}
	return &food, nil
}

func (r *foodRepository) FindCustomFoodByName(name string) (*models.CustomFood, error) {
	var food models.CustomFood
	err := r.customFoodCollection.FindOne(context.TODO(), bson.M{"name": name}).Decode(&food)
	if err != nil {
		return nil, err
	}
	return &food, nil
}

func (r *foodRepository) GetAllStandardFoods() ([]*models.StandardFood, error) {
	var foods []*models.StandardFood

	filter := bson.M{}
	cursor, err := r.standardFoodCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &foods); err != nil {
		return nil, err
	}

	return foods, nil
}

func (r *foodRepository) GetAllCustomFoods() ([]*models.CustomFood, error) {
	var foods []*models.CustomFood

	filter := bson.M{}
	cursor, err := r.customFoodCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &foods); err != nil {
		return nil, err
	}

	return foods, nil
}

func (r *foodRepository) SaveStandardFood(food *models.StandardFood) error {
	_, err := r.standardFoodCollection.InsertOne(context.TODO(), food)
	return err
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

func (r *foodRepository) UpdateCreatedReviewStats(foodIDs []primitive.ObjectID, rating int) error {
	if len(foodIDs) == 0 {
		return nil
	}
	if rating <= 0 || rating > 5 {
		return nil
	}
	
	filter := bson.M{"_id": bson.M{"$in": foodIDs}}

	incMap := bson.M{"reviewCount": 1}
	incMap["totalRating"] = rating

	update := bson.M{"$inc": incMap}
	_, err := r.standardFoodCollection.UpdateMany(context.TODO(), filter, update)
	return err
}

func (r *foodRepository) UpdateModifiedReviewStats(foodIDs []primitive.ObjectID, oldRating, newRating int) error {
	if len(foodIDs) == 0 {
		return nil
	}
	if newRating <= 0 || newRating > 5 {
		return nil
	}
	
	filter := bson.M{"_id": bson.M{"$in": foodIDs}}

	ratingDiff := newRating - oldRating

	incMap := bson.M{}
	if ratingDiff != 0 {
		incMap["totalRating"] = ratingDiff
	} else {
		return nil
	}

	update := bson.M{"$inc": incMap}
	_, err := r.standardFoodCollection.UpdateMany(context.TODO(), filter, update)
	return err
}

func (r *foodRepository) IncrementLikeCount(foodID primitive.ObjectID) error {
	filter := bson.M{"_id": foodID}
	update := bson.M{"$inc": bson.M{"likeCount": 1}}
	_, err := r.standardFoodCollection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *foodRepository) DecrementLikeCount(foodID primitive.ObjectID) error {
	filter := bson.M{"_id": foodID}
	update := bson.M{"$inc": bson.M{"likeCount": -1}}
	_, err := r.standardFoodCollection.UpdateOne(context.TODO(), filter, update)
	return err
}