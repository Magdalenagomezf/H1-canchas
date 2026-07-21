package mercadopago

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
)

func TestClientGetPayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/v1/payments/987654321" {
			t.Errorf("path = %q, want /v1/payments/987654321", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payment.Response{
			ID:                987654321,
			Status:            "approved",
			StatusDetail:      "accredited",
			ExternalReference: "booking-42",
			TransactionAmount: 1500.5,
		})
	}))
	defer server.Close()

	cfg, err := config.New("TEST-TOKEN", config.WithHTTPClient(newTestRequester(t, server)))
	if err != nil {
		t.Fatalf("config.New() error = %v", err)
	}

	client := &Client{payment: payment.NewClient(cfg)}

	got, err := client.GetPayment(context.Background(), 987654321)
	if err != nil {
		t.Fatalf("GetPayment() error = %v", err)
	}

	want := &PaymentInfo{
		ID:                987654321,
		Status:            "approved",
		StatusDetail:      "accredited",
		ExternalReference: "booking-42",
		TransactionAmount: 1500.5,
	}

	if *got != *want {
		t.Errorf("GetPayment() = %+v, want %+v", *got, *want)
	}
}

func TestClientGetPaymentError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"payment not found"}`))
	}))
	defer server.Close()

	cfg, err := config.New("TEST-TOKEN", config.WithHTTPClient(newTestRequester(t, server)))
	if err != nil {
		t.Fatalf("config.New() error = %v", err)
	}

	client := &Client{payment: payment.NewClient(cfg)}

	_, err = client.GetPayment(context.Background(), 1)
	if err == nil {
		t.Fatal("GetPayment() error = nil, want error on non-2xx response")
	}
}
