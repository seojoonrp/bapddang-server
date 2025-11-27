// api/repositories/review_repository.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReviewRepository interface {
	SaveReview(review *models.Review) error
	UpdateReview(review *models.Review) error
	FindByUserIDAndDay(userID primitive.ObjectID, day int) ([]models.Review, error)
	FindByIDAndUserID(reviewID, userID primitive.ObjectID) (*models.Review, error)
}

type reviewRepository struct {
	collection *mongo.Collection
}

func NewReviewRepository(coll *mongo.Collection) ReviewRepository {
	return &reviewRepository{collection: coll}
}

func (r *reviewRepository) SaveReview (review *models.Review) error {
	_, err := r.collection.InsertOne(context.TODO(), review)
	return err
}

func (r *reviewRepository) UpdateReview (review *models.Review) error {
	filter := bson.M{"_id": review.ID}
	update := bson.M{
		"$set": bson.M{
			"mealTime": review.MealTime,
			"tags": review.Tags,
			"imageURL": review.ImageURL,
			"comment": review.Comment,
			"rating": review.Rating,
			"updatedAt": review.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *reviewRepository) FindByUserIDAndDay(userID primitive.ObjectID, day int) ([]models.Review, error) {
	var reviews []models.Review

	filter := bson.M{"userId": userID, "day": day}
	cursor, err := r.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &reviews); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *reviewRepository) FindByIDAndUserID(reviewID, userID primitive.ObjectID) (*models.Review, error) {
	var review models.Review

	filter := bson.M{"_id": reviewID, "userId": userID}
	err := r.collection.FindOne(context.TODO(), filter).Decode(&review)
	if err != nil {
		return nil, err
	}
	
	return &review, nil
}