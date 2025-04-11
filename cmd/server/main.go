package main

import (
	"log"
	"process-payments/internal/config"
	"process-payments/internal/database"
	"process-payments/internal/repository"
	"process-payments/internal/server"
	"process-payments/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Load config
	cfg := config.GetConfig()

	// Load database
	cfg.MongoClient = database.DBInstance(cfg)

	//Initialize collections
	cfg.Collections = &repository.Collections{
		PaymentCollection: repository.NewMongoPaymentRepository(database.OpenCollection(cfg.MongoClient, "transactions")),
	}

	//Initialize Services
	cfg.Services = &services.Services{
		StripeService: services.NewStripeService(cfg.ENV.STRIPE_SECRET_KEY, cfg.ENV.STRIPE_WEBHOOK_SECRET_KEY, cfg.Products, cfg.Production, cfg.Collections),
	}

	cfg.UpdateConfig()

	// Start server
	server.StartServer(cfg)
}
