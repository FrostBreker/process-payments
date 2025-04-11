package types

type StripeCheckoutRequest struct {
	ProductId string
	UserId    string
	ReturnURL string
}
