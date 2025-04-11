package controllers

import (
	"log"
	"process-payments/internal/config"
	"process-payments/internal/utils"
	"process-payments/pkg/types"

	"github.com/gin-gonic/gin"
)

// HandleStripeWebhooks The `HandleStripeWebhooks` function is a controller that handles webhook requests from Stripe.
func HandleStripeWebhooks() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()

		stripeService := cfg.Services.StripeService
		event, err := stripeService.AuthenticateWebhook(c)
		if err == nil {
			go func() {
				err := stripeService.HandleEvents(event)
				if err != nil {
					log.Println("Error handling stripe event: ", err.Error())
				}
			}()
		} else {
			utils.SendResponse(c, false, 400, err.Error(), "Webhook is not valid", nil)
		}
		utils.SendResponse(c, true, 200, "", "Webhook received successfully", nil)
	}
}

// CreateStripeCheckout The `GetStripeCheckout` function is a controller that handles requests to create the Stripe checkout session.
func CreateStripeCheckout() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()
		productId := c.Query("productId")
		userId := c.GetString("userId")

		if productId == "" {
			utils.SendResponse(c, false, 400, "productId is required", "Error getting checkout session", nil)
			return
		}
		if userId == "" {
			utils.SendResponse(c, false, 400, "userId is required", "Error getting checkout session", nil)
			return
		}

		//Construct StripeCheckoutRequest
		stripeCheckoutRequest := types.StripeCheckoutRequest{
			UserId:    userId,
			ProductId: productId,
			ReturnURL: cfg.ClientURL + "/account",
		}

		stripeService := cfg.Services.StripeService
		session, err := stripeService.GetCheckoutSession(stripeCheckoutRequest)
		if err != nil {
			utils.SendResponse(c, false, 400, "Error getting checkout session", "Error getting checkout session", nil)
			return
		}
		utils.SendResponse(c, true, 200, "", "Checkout session retrieved successfully", gin.H{
			"sessionId": session.ID,
			"url":       string(session.URL),
		})
	}
}
