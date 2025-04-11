package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	SubscriptionID string             `bson:"subscriptionId"`
	User           UserInSubscription `bson:"user"`
	Plan           PlanInSubscription `bson:"plan"`
	InvoiceLink    string             `bson:"invoiceLink"`
	InvoicePDF     string             `bson:"invoicePDF"`
	InvoiceNumber  string             `bson:"invoiceNumber"`
	IsTest         bool               `bson:"isTest"`
	IsOneTime      bool               `bson:"isOneTime"`
	IsCanceled     bool               `bson:"isCanceled"`
	UserId         string             `bson:"userId"`
	Status         string             `bson:"status"`
	EndsAt         int64              `bson:"endsAt"`
	CreatedAt      int64              `bson:"createdAt"`
	UpdatedAt      int64              `bson:"updatedAt"`
	RenewsAt       int64              `bson:"renewsAt"`
}

type UserInSubscription struct {
	Email      string `bson:"email"`
	Name       string `bson:"name"`
	CustomerId string `bson:"customerId"`
}

type PlanInSubscription struct {
	SessionId string  `bson:"sessionId"`
	ProductId string  `bson:"productId"`
	Price     float32 `bson:"price"`
}
