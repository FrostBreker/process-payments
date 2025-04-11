package routes

import (
	"process-payments/internal/controllers"

	"github.com/gin-gonic/gin"
)

// StripeRoutes The `StripeRoutes` function sets up webhook routes for logging in and generating authentication
// URLs.
func StripeRoutes(router *gin.RouterGroup) {
	router.Use(func(c *gin.Context) {
		c.Set("userId", "123")
		c.Next()
	})

	// Webhooks
	router.POST("/webhooks", controllers.HandleStripeWebhooks())

	// Checkout
	router.GET("/", controllers.CreateStripeCheckout())
}
