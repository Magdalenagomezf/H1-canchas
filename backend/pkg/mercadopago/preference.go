package mercadopago

import (
	"context"

	"github.com/mercadopago/sdk-go/pkg/preference"
)

type BackURLs struct {
	Success string
	Pending string
	Failure string
}

type CreatePreferenceRequest struct {
	Title             string
	Amount            float64
	ExternalReference string
	BackURLs          BackURLs
	NotificationURL   string
}

type PreferenceResponse struct {
	ID               string
	InitPoint        string
	SandboxInitPoint string
}

func (c *Client) CreatePreference(ctx context.Context, req CreatePreferenceRequest) (*PreferenceResponse, error) {
	resp, err := c.preference.Create(ctx, buildPreferenceRequest(req))
	if err != nil {
		return nil, err
	}

	return &PreferenceResponse{
		ID:               resp.ID,
		InitPoint:        resp.InitPoint,
		SandboxInitPoint: resp.SandboxInitPoint,
	}, nil
}

func buildPreferenceRequest(req CreatePreferenceRequest) preference.Request {
	return preference.Request{
		Items: []preference.ItemRequest{
			{
				Title:     req.Title,
				UnitPrice: req.Amount,
				Quantity:  1,
			},
		},
		BackURLs: &preference.BackURLsRequest{
			Success: req.BackURLs.Success,
			Pending: req.BackURLs.Pending,
			Failure: req.BackURLs.Failure,
		},
		ExternalReference: req.ExternalReference,
		NotificationURL:   req.NotificationURL,
	}
}
