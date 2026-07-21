package mercadopago

import "context"

type PaymentInfo struct {
	ID                int64
	Status            string
	StatusDetail      string
	ExternalReference string
	TransactionAmount float64
}

func (c *Client) GetPayment(ctx context.Context, paymentID int64) (*PaymentInfo, error) {
	resp, err := c.payment.Get(ctx, int(paymentID))
	if err != nil {
		return nil, err
	}

	return &PaymentInfo{
		ID:                int64(resp.ID),
		Status:            resp.Status,
		StatusDetail:      resp.StatusDetail,
		ExternalReference: resp.ExternalReference,
		TransactionAmount: resp.TransactionAmount,
	}, nil
}
