package repository

import (
	"context"
	"errors"
	"log"
	"process-payments/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PaymentRepository interface {
	Save(subs *models.Subscription) error
	GetByUserId(userId string) (*models.Subscription, error)
	Get(subscriptionId string) (*models.Subscription, error)
	Update(subscription *models.Subscription) error
	Delete(subscriptionId string) error
	// IsValid checks if a subscription is valid for a given user ID
	IsValid(userId string) bool
}

type MongoPaymentRepository struct {
	collection *mongo.Collection
}

// Errors
var (
	ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrorUpdatingSubscription    = errors.New("error updating subscription")
	ErrorDeletingSubscription    = errors.New("error deleting subscription")
)

func NewMongoPaymentRepository(collection *mongo.Collection) PaymentRepository {
	return &MongoPaymentRepository{collection: collection}
}

// Save a subscription object into the database
func (r *MongoPaymentRepository) Save(subs *models.Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//check if userId has already a subscription object
	_, err := r.Get(subs.SubscriptionID)
	if err == nil {
		return ErrSubscriptionAlreadyExists
	}

	_, err = r.collection.InsertOne(ctx, subs)
	if err != nil {
		return err
	}
	return nil
}

// Get a subscription by subscriptionId
func (r *MongoPaymentRepository) Get(subscriptionId string) (*models.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var subscription models.Subscription

	filter := bson.M{"subscriptionId": subscriptionId}

	err := r.collection.FindOne(ctx, filter).Decode(&subscription)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrSubscriptionNotFound // Subscription not found, return nil subscription
		}
		log.Printf("Error finding subscription: %v", err)
		return nil, err // An error occurred
	}

	return &subscription, nil
}

// GetByUserId a subscription by userId
func (r *MongoPaymentRepository) GetByUserId(userId string) (*models.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var subscription models.Subscription
	filter := bson.M{"userId": userId}

	err := r.collection.FindOne(ctx, filter).Decode(&subscription)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrSubscriptionNotFound // Subscription not found, return nil subscription
		}
		log.Printf("Error finding subscription: %v", err)
		return nil, err // An error occurred
	}

	return &subscription, nil
}

// Update subscription
func (r *MongoPaymentRepository) Update(subscription *models.Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"subscriptionId": subscription.SubscriptionID}

	_, err := r.collection.ReplaceOne(ctx, filter, subscription)
	if err != nil {
		return ErrorUpdatingSubscription
	}

	return nil
}

// Delete a subscription
func (r *MongoPaymentRepository) Delete(subscriptionId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"subscriptionId": subscriptionId}

	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrSubscriptionNotFound
		} else {
			return ErrorDeletingSubscription
		}
	}

	return nil
}

// IsValid Check if a subscription is valid
func (r *MongoPaymentRepository) IsValid(userId string) bool {

	subs, err := r.GetByUserId(userId)
	if err != nil {
		return false
	}

	var isPremium bool

	status := subs.Status
	expiresAt := subs.EndsAt

	if subs.IsOneTime && (status == "active" || status == "paid" || status == "complete") && expiresAt == -1 {
		return true
	}

	// If the subscription is active and not expired, the user has premium access
	isPremium = (status == "active" || status == "trialing") && time.Now().Before(time.Unix(expiresAt/1000, 0))
	return isPremium
}
