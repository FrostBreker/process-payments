package services

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"process-payments/internal/models"
	"process-payments/internal/repository"
	"process-payments/pkg/types"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/invoice"
	"github.com/stripe/stripe-go/v82/product"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

const TwelveHoursInMilliseconds int64 = 3600 * 12 * 1000

type StripeService struct {
	stripeSecretKey        string
	stripeWebhookSecretKey string
	products               []string
	isProd                 bool
	repo                   *repository.Collections
}

// NewStripeService creates a new instance of the StripeService
func NewStripeService(stripeSecretKey, stripeWebhookSecretKey string, products []string, prod bool, collection *repository.Collections) *StripeService {
	stripe.Key = stripeSecretKey
	return &StripeService{
		stripeSecretKey:        stripeSecretKey,
		stripeWebhookSecretKey: stripeWebhookSecretKey,
		products:               products,
		isProd:                 prod,
		repo:                   collection,
	}
}

// AllowedStripeIPs are the IPs from which Stripe sends webhooks
var AllowedStripeIPs = []string{
	"3.18.12.63",
	"3.130.192.231",
	"13.235.14.237",
	"13.235.122.149",
	"18.211.135.69",
	"35.154.171.200",
	"52.15.183.38",
	"54.88.130.119",
	"54.88.130.237",
	"54.187.174.169",
	"54.187.205.235",
	"54.187.216.72",
}

// Authenticating errors
var (
	ErrWebhookNotFromStripe = errors.New("webhook is not coming from stripe")
	ErrReadingRequestBody   = errors.New("error reading request body")
	ErrorVerifyingSignature = errors.New("error verifying webhook signature")
)

// Handling Webhook errors
var (
	ErrorHandlingStripeEvent = errors.New("error handling stripe event")
	ErrParsingWebhookJSON    = errors.New("error parsing webhook JSON")
	ErrSubscriptionNotFound  = errors.New("subscription not found")
	ErrCustomUserIdNotExist  = errors.New("custom user id does not exist")
)

// Handling Invoices Error
var (
	ErrGettingInvoice = errors.New("error getting invoice")
)

// Handling subscriptions Errors
var (
	ErrGettingSubscription = errors.New("error getting subscription")
)

// Handling customers errors
var (
	ErrCreatingCustomer = errors.New("error creating customer")
	ErrGettingCustomer  = errors.New("error getting customer")
)

// Handling products errors
var (
	ErrorGettingProduct = errors.New("error getting product")
)

// Handling checkout creation errors
var (
	ErrorCreatingCheckout = errors.New("error creating checkout session")
)

// CreateCustomer creates a new customer in Stripe
func (s *StripeService) CreateCustomer(userId string) (*stripe.Customer, error) {
	customerParams := &stripe.CustomerParams{
		Metadata: map[string]string{
			"userId": userId,
		},
		Address: &stripe.AddressParams{
			Country: stripe.String("FR"),
		},
	}
	customerData, err := customer.New(customerParams)
	if err != nil {
		log.Printf("Error creating customer: %v", err)
		return nil, ErrCreatingCustomer
	}
	return customerData, nil
}

// GetCustomer retrieves a customer from Stripe
func (s *StripeService) GetCustomer(customerId string) (*stripe.Customer, error) {
	customerParams := &stripe.CustomerParams{}
	customerData, err := customer.Get(customerId, customerParams)
	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return nil, ErrGettingCustomer
	}
	return customerData, nil
}

// GetCustomerByUserId retrieves a customer from Stripe
func (s *StripeService) GetCustomerByUserId(userId string) (*stripe.Customer, error) {

	params := &stripe.CustomerSearchParams{
		SearchParams: stripe.SearchParams{
			Query: "metadata['userId']:'" + userId + "'",
		},
	}
	result := customer.Search(params)
	customers := result.CustomerSearchResult().Data
	if len(customers) < 1 {
		log.Printf("Error getting customer: %v", ErrGettingCustomer)
		return nil, ErrGettingCustomer
	}

	if customers[0].Metadata["userId"] != userId {
		log.Printf("Error getting customer: %v", ErrGettingCustomer)
		return nil, ErrGettingCustomer
	}

	return customers[0], nil
}

//Products

// GetProduct retrieves a product from Stripe
func (s *StripeService) GetProduct(productId string) (*stripe.Product, error) {
	params := &stripe.ProductParams{}
	productData, err := product.Get(productId, params)

	if err != nil {
		log.Printf("Error getting product: %v", err)
		return nil, ErrorGettingProduct
	}

	return productData, nil
}

//Invoices

// GetInvoice retrieves an invoice from Stripe
func (s *StripeService) GetInvoice(invoiceId string) (*stripe.Invoice, error) {
	params := &stripe.InvoiceParams{}
	invoiceData, err := invoice.Get(invoiceId, params)

	if err != nil {
		log.Printf("Error getting invoice: %v", err)
		return nil, ErrGettingInvoice
	}

	return invoiceData, nil
}

//Subscriptions

// GetSubscription retrieves a subscription from Stripe
func (s *StripeService) GetSubscription(subscriptionId string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{}
	subscriptionData, err := subscription.Get(subscriptionId, params)

	if err != nil {
		log.Printf("Error getting subscription: %v", err)
		return nil, ErrGettingSubscription
	}

	return subscriptionData, nil
}

//Webhooks

// AuthenticateWebhook authenticates the webhook request
func (s *StripeService) AuthenticateWebhook(c *gin.Context) (stripe.Event, error) {
	if !s.isProd {
		AllowedStripeIPs = append(AllowedStripeIPs, "::1")
	}
	req := c.Request
	ipFromStripe := c.ClientIP()

	// Checks webhook coming from allowed IP
	if !slices.Contains[[]string](AllowedStripeIPs[:], ipFromStripe) {
		log.Printf("Error authenticating webhook: %v", ErrWebhookNotFromStripe)
		return stripe.Event{}, ErrWebhookNotFromStripe
	}

	// Checks webhook signature.
	// This makes sure that the POST request is actually coming from Stripe.
	signatureHeader := req.Header.Get("Stripe-Signature")
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(c.Writer, req.Body, MaxBodyBytes)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		return stripe.Event{}, ErrReadingRequestBody
	}

	event, err := webhook.ConstructEvent(body, signatureHeader, s.stripeWebhookSecretKey)
	if err != nil {
		log.Printf("Error verifying signature: %v", err)
		return stripe.Event{}, ErrorVerifyingSignature
	}

	return event, nil
}

// HandleEvents from webhooks
func (s *StripeService) HandleEvents(e stripe.Event) error {
	switch e.Type {

	case "customer.subscription.updated":
		var customerSubscription stripe.Subscription
		err := json.Unmarshal(e.Data.Raw, &customerSubscription)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			return ErrParsingWebhookJSON
		}

		if customerSubscription.ID != "" {
			err = s.handleSubscriptionUpdate(customerSubscription)
			if err != nil {
				log.Printf("Error handling subscription update: %v", err)
				return err
			}
		}
	case "customer.subscription.deleted":
		var customerSubscription stripe.Subscription
		err := json.Unmarshal(e.Data.Raw, &customerSubscription)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			return ErrParsingWebhookJSON
		}

		if customerSubscription.ID != "" {
			err = s.handleSubscriptionCancellation(customerSubscription)
			if err != nil {
				log.Printf("Error handling subscription cancellation: %v", err)
				return err
			}
		} else {
			log.Printf("Error handling subscription cancellation: %v", ErrSubscriptionNotFound)
			return ErrSubscriptionNotFound
		}
		return nil
	case "checkout.session.completed":
		var sessionData stripe.CheckoutSession
		err := json.Unmarshal(e.Data.Raw, &sessionData)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			return ErrParsingWebhookJSON
		}

		// Handle payment completion
		if sessionData.Mode == stripe.CheckoutSessionModeSubscription {
			return s.handleSubscriptionPaymentCompletion(sessionData)
		}
	default:
		log.Printf("Error handling stripe event: %v", ErrorHandlingStripeEvent)
		return ErrorHandlingStripeEvent
	}
	return nil
}

// handleSubscriptionPaymentCompletion handles the completion of a subscription payment
func (s *StripeService) handleSubscriptionPaymentCompletion(checkoutSession stripe.CheckoutSession) error {
	customerData, err := s.GetCustomer(checkoutSession.Customer.ID)
	if err != nil {
		return err
	}

	customerUserId := checkoutSession.ClientReferenceID
	if customerUserId == "" {
		log.Printf("Error handling one-time payment completion: %v", ErrCustomUserIdNotExist)
		return ErrCustomUserIdNotExist
	}

	invoiceData, err := s.GetInvoice(checkoutSession.Invoice.ID)
	if err != nil {
		return err
	}

	subscriptionData, err := s.GetSubscription(checkoutSession.Subscription.ID)
	if err != nil {
		return err
	}

	subscriptionStatus := subscriptionData.Status // Possible values are `incomplete`, `incomplete_expired`, `trialing`, `active`, `past_due`, `canceled`, or `unpaid`.
	expireDateTimestamp := subscriptionData.Items.Data[0].CurrentPeriodEnd * 1000
	// Add 12h as a security. Sometimes the invoice takes some time to be processed even when there's nothing wrong with the payment methods.
	expireDateTimestamp += TwelveHoursInMilliseconds

	subscriptionModel := &models.Subscription{
		UserId:         customerUserId,
		SubscriptionID: subscriptionData.ID,
		InvoiceLink:    invoiceData.HostedInvoiceURL,
		InvoicePDF:     invoiceData.InvoicePDF,
		InvoiceNumber:  invoiceData.Number,
		IsTest:         invoiceData.Livemode,
		IsOneTime:      false,
		Status:         string(subscriptionStatus),
		EndsAt:         expireDateTimestamp,
		CreatedAt:      subscriptionData.Created * 1000,
		IsCanceled:     false,
		RenewsAt:       expireDateTimestamp,
		UpdatedAt:      time.Now().UnixMilli(),
		User: models.UserInSubscription{
			Email:      customerData.Email,
			Name:       customerData.Name,
			CustomerId: customerData.ID,
		},
		Plan: models.PlanInSubscription{
			SessionId: checkoutSession.ID,
			ProductId: subscriptionData.Items.Data[0].Price.Product.ID,
			Price:     float32(invoiceData.Total / 100),
		},
	}

	err = s.repo.PaymentCollection.Save(subscriptionModel)
	if err != nil {
		err := s.repo.PaymentCollection.Update(subscriptionModel)
		if err != nil {
			log.Printf("Error updating payment: %v", err)
			return err
		}
	}

	return nil
}

// handleSubscriptionUpdate handles the update of a subscription
func (s *StripeService) handleSubscriptionUpdate(subscription stripe.Subscription) error {
	subscriptionStatus := subscription.Status // Possible values are `incomplete`, `incomplete_expired`, `trialing`, `active`, `past_due`, `canceled`, or `unpaid`.
	expireDateTimestamp := subscription.Items.Data[0].CurrentPeriodEnd * 1000
	// Add 12h as a security. Sometimes the invoice takes some time to be processed even when there's nothing wrong with the payment methods.
	expireDateTimestamp += TwelveHoursInMilliseconds
	// get invoice data
	invoiceData, err := s.GetInvoice(subscription.LatestInvoice.ID)
	if err != nil {
		return err
	}

	// update user subscription status and expire date
	subscriptionData, err := s.repo.PaymentCollection.Get(subscription.ID)
	if err != nil {
		log.Printf("Error getting subscription: %v", err)
		return ErrSubscriptionNotFound
	}
	subscriptionData.Plan = models.PlanInSubscription{
		ProductId: subscription.Items.Data[0].Price.Product.ID,
		Price:     float32(invoiceData.Total / 100),
		SessionId: subscriptionData.Plan.SessionId,
	}
	subscriptionData.InvoicePDF = invoiceData.InvoicePDF
	subscriptionData.InvoiceLink = invoiceData.HostedInvoiceURL
	subscriptionData.InvoiceNumber = invoiceData.Number
	subscriptionData.Status = string(subscriptionStatus)
	subscriptionData.EndsAt = expireDateTimestamp
	subscriptionData.RenewsAt = expireDateTimestamp
	subscriptionData.UpdatedAt = time.Now().UnixMilli()
	subscriptionData.IsCanceled = subscription.CancelAtPeriodEnd

	err = s.repo.PaymentCollection.Update(subscriptionData)
	if err != nil {
		log.Printf("Error updating payment: %v", err)
		return err
	}

	return nil
}

// handleSubscriptionCancellation handles the cancellation of a subscription
func (s *StripeService) handleSubscriptionCancellation(subscription stripe.Subscription) error {

	_, err := s.repo.PaymentCollection.Get(subscription.ID)
	if err != nil {
		log.Printf("Error getting subscription: %v", err)
		return ErrSubscriptionNotFound
	}

	err = s.repo.PaymentCollection.Delete(subscription.ID)
	if err != nil {
		if errors.Is(err, repository.ErrorDeletingSubscription) {
			log.Printf("Error deleting subscription: %v", err)
			return err
		}
	}

	return nil
}

// Checkouts

// GetCheckoutSession returns the Stripe checkout session
func (s *StripeService) GetCheckoutSession(request types.StripeCheckoutRequest) (*stripe.CheckoutSession, error) {

	//Get the product data
	productData, err := s.GetProduct(request.ProductId)
	if err != nil {
		return nil, err
	}

	//Get the checkout mode by checking if the product is a subscription or not
	var isSubscription bool
	var checkoutMode stripe.CheckoutSessionMode
	isSubscription = productData.Metadata["subs"] == "true"
	if isSubscription {
		checkoutMode = stripe.CheckoutSessionModeSubscription
	} else {
		checkoutMode = stripe.CheckoutSessionModePayment
	}
	isTrial := productData.Metadata["trial"] == "true"

	checkoutParams := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(checkoutMode)),
		SuccessURL:        stripe.String(request.ReturnURL),
		CancelURL:         stripe.String(request.ReturnURL),
		ClientReferenceID: stripe.String(request.UserId),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(productData.DefaultPrice.ID),
				Quantity: stripe.Int64(1),
			},
		},
	}

	if isSubscription {
		if isTrial {
			checkoutParams.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
				TrialPeriodDays: stripe.Int64(14),
			}
		}
	}

	sessionData, err := session.New(checkoutParams)
	if err != nil {
		log.Printf("Error creating checkout: %v", err)
		return nil, ErrorCreatingCheckout
	}

	return sessionData, nil
}
