package mercadopago

import "testing"

func TestNewClient(t *testing.T) {
	client, err := NewClient("TEST-TOKEN")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}
	if client.preference == nil {
		t.Error("client.preference is nil")
	}
	if client.payment == nil {
		t.Error("client.payment is nil")
	}
}
