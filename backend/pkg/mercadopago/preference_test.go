package mercadopago

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

func TestBuildPreferenceRequest(t *testing.T) {
	req := CreatePreferenceRequest{
		Title:             "Cancha 1 - Turno 18:00",
		Amount:            1500.5,
		ExternalReference: "booking-42",
		BackURLs: BackURLs{
			Success: "https://example.com/success",
			Pending: "https://example.com/pending",
			Failure: "https://example.com/failure",
		},
	}

	got := buildPreferenceRequest(req)

	if got.NotificationURL != "" {
		t.Errorf("NotificationURL = %q, want empty (single dashboard webhook, not per-preference)", got.NotificationURL)
	}

	if got.ExternalReference != req.ExternalReference {
		t.Errorf("ExternalReference = %q, want %q", got.ExternalReference, req.ExternalReference)
	}

	if len(got.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(got.Items))
	}

	item := got.Items[0]
	if item.Title != req.Title {
		t.Errorf("Items[0].Title = %q, want %q", item.Title, req.Title)
	}
	if item.UnitPrice != req.Amount {
		t.Errorf("Items[0].UnitPrice = %v, want %v", item.UnitPrice, req.Amount)
	}
	if item.Quantity != 1 {
		t.Errorf("Items[0].Quantity = %d, want 1", item.Quantity)
	}

	if got.BackURLs == nil {
		t.Fatal("BackURLs is nil")
	}
	if got.BackURLs.Success != req.BackURLs.Success {
		t.Errorf("BackURLs.Success = %q, want %q", got.BackURLs.Success, req.BackURLs.Success)
	}
	if got.BackURLs.Pending != req.BackURLs.Pending {
		t.Errorf("BackURLs.Pending = %q, want %q", got.BackURLs.Pending, req.BackURLs.Pending)
	}
	if got.BackURLs.Failure != req.BackURLs.Failure {
		t.Errorf("BackURLs.Failure = %q, want %q", got.BackURLs.Failure, req.BackURLs.Failure)
	}
}

func TestClientCreatePreference(t *testing.T) {
	var capturedBody preference.Request

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/checkout/preferences" {
			t.Errorf("path = %q, want /checkout/preferences", r.URL.Path)
		}

		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(preference.Response{
			ID:               "pref-123",
			InitPoint:        "https://mercadopago.com/checkout/init/pref-123",
			SandboxInitPoint: "https://sandbox.mercadopago.com/checkout/init/pref-123",
		})
	}))
	defer server.Close()

	cfg, err := config.New("TEST-TOKEN", config.WithHTTPClient(newTestRequester(t, server)))
	if err != nil {
		t.Fatalf("config.New() error = %v", err)
	}

	client := &Client{
		preference: preference.NewClient(cfg),
		payment:    nil,
	}

	req := CreatePreferenceRequest{
		Title:             "Cancha 2 - Turno 20:00",
		Amount:            2000,
		ExternalReference: "booking-99",
		BackURLs: BackURLs{
			Success: "https://example.com/success",
			Pending: "https://example.com/pending",
			Failure: "https://example.com/failure",
		},
	}

	got, err := client.CreatePreference(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePreference() error = %v", err)
	}

	if got.ID != "pref-123" {
		t.Errorf("ID = %q, want pref-123", got.ID)
	}
	if got.InitPoint != "https://mercadopago.com/checkout/init/pref-123" {
		t.Errorf("InitPoint = %q, want the production init point", got.InitPoint)
	}
	if got.SandboxInitPoint != "https://sandbox.mercadopago.com/checkout/init/pref-123" {
		t.Errorf("SandboxInitPoint = %q, want the sandbox init point", got.SandboxInitPoint)
	}

	if capturedBody.ExternalReference != req.ExternalReference {
		t.Errorf("captured ExternalReference = %q, want %q", capturedBody.ExternalReference, req.ExternalReference)
	}
	if capturedBody.NotificationURL != "" {
		t.Errorf("captured NotificationURL = %q, want empty", capturedBody.NotificationURL)
	}
}

func TestClientCreatePreferenceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"invalid preference"}`))
	}))
	defer server.Close()

	cfg, err := config.New("TEST-TOKEN", config.WithHTTPClient(newTestRequester(t, server)))
	if err != nil {
		t.Fatalf("config.New() error = %v", err)
	}

	client := &Client{preference: preference.NewClient(cfg)}

	_, err = client.CreatePreference(context.Background(), CreatePreferenceRequest{Title: "x", Amount: 1})
	if err == nil {
		t.Fatal("CreatePreference() error = nil, want error on non-2xx response")
	}
}
