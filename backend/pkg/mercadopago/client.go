// Package mercadopago wraps the official Mercado Pago Go SDK behind a small,
// project-specific API surface (Checkout Pro preferences, payment lookups,
// and webhook signature verification).
package mercadopago

import (
	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

type Client struct {
	preference preference.Client
	payment    payment.Client
}

func NewClient(accessToken string) (*Client, error) {
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, err
	}

	return &Client{
		preference: preference.NewClient(cfg),
		payment:    payment.NewClient(cfg),
	}, nil
}
