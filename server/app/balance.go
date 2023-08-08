// Package app for c4s backend app
package app

import (
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
)

// CreateBalanceProductInStripe creates a new stripe product for balance
func createBalanceProductInStripe(balance int64) (string, error) {
	params := &stripe.ProductParams{Name: stripe.String("user balance")}
	prod, err := product.New(params)
	if err != nil {
		return "", err
	}

	paramsPrice := &stripe.PriceParams{
		Product:    stripe.String(prod.ID),
		UnitAmount: stripe.Int64(balance),
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
	}

	priceObj, err := price.New(paramsPrice)
	if err != nil {
		return "", err
	}

	return priceObj.ID, nil
}
