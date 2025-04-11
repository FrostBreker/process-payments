package middlewares

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(clientURL string, prod bool) gin.HandlerFunc {
	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Authentication", "Stripe-Signature"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if prod {
		config.AllowOrigins = []string{clientURL}
	} else {
		config.AllowOrigins = []string{"http://127.0.0.1:3000", "http://localhost:3000"}
	}

	return cors.New(config)
}
