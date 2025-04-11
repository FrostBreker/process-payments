package config

import (
	"log"
	"os"
	"process-payments/internal/repository"
	"process-payments/internal/services"
	"strconv"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	AppName     string
	Production  bool
	Port        string
	ClientURL   string
	ENV         ENV
	MongoClient *mongo.Client
	Collections *repository.Collections
	Services    *services.Services
	Products    []string
}

type ENV struct {
	PORT                      string
	MONGO_URI                 string
	PRODUCTION                bool
	STRIPE_WEBHOOK_SECRET_KEY string
	STRIPE_SECRET_KEY         string
}

var configInstance *Config
var once sync.Once

// The function `ConvertStringToBool` converts a string to a boolean value in Go.
func convertStringToBool(str string) bool {
	value, err := strconv.ParseBool(str)
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func LoadConfig() *Config {
	prod := convertStringToBool(os.Getenv("PRODUCTION"))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	configInstance = &Config{
		AppName:    "ProcessPaymentsAPI",
		Production: prod,
		Port:       port,
		ClientURL:  os.Getenv("CLIENT_URL"),
		ENV: ENV{
			PORT:                      port,
			MONGO_URI:                 os.Getenv("MONGO_URI"),                 // MongoDB URI
			PRODUCTION:                prod,                                   // Production flag
			STRIPE_WEBHOOK_SECRET_KEY: os.Getenv("STRIPE_WEBHOOK_SECRET_KEY"), // Stripe Webhook Secret Key
			STRIPE_SECRET_KEY:         os.Getenv("STRIPE_SECRET_KEY"),         // Stripe Secret Key
		},
		Products: []string{"prod_S6WxyFWfWVsP60"},
	}

	return configInstance
}

func GetConfig() *Config {
	once.Do(func() {
		configInstance = LoadConfig()
	})
	return configInstance
}

func (c *Config) UpdateConfig() {
	configInstance = c
}
