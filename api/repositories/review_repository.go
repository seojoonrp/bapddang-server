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
	FindByUserIDAndDay(userID primitive.ObjectID, day int) ([]models.Review, error)
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