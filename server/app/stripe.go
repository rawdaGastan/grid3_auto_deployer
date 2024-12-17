package app

import (
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/paymentmethod"
)

func createCustomer(name, email string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Name:  stripe.String(name),
		Email: stripe.String(email),
	}

	return customer.New(params)
}

func createPaymentIntent(customerID, paymentMethodID, currency string, amount float64) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(amount * 100)),
		Currency:      stripe.String(currency),
		Customer:      stripe.String(customerID),
		PaymentMethod: stripe.String(paymentMethodID),
		Confirm:       stripe.Bool(true), // Automatically confirm the payment
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled:        stripe.Bool(true),
			AllowRedirects: stripe.String("never"),
		},
	}

	return paymentintent.New(params)
}

func createPaymentMethod(cardType, paymentMethodID string) (*stripe.PaymentMethod, error) {
	paymentMethodParams := &stripe.PaymentMethodParams{
		Type: stripe.String(cardType),
		Card: &stripe.PaymentMethodCardParams{Token: stripe.String(paymentMethodID)},
	}

	return paymentmethod.New(paymentMethodParams)
}

func attachPaymentMethod(customerID, paymentMethodID string) error {
	paymentMethodAttachParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}

	_, err := paymentmethod.Attach(paymentMethodID, paymentMethodAttachParams)
	return err
}

func detachPaymentMethod(paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	return err
}

func updateDefaultPaymentMethod(customerID, paymentMethodID string) error {
	_, err := customer.Update(customerID, &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	})
	return err
}
