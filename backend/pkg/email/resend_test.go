package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResendSender_Send_RequestShape(t *testing.T) {
	var gotMethod, gotPath, gotAuth, gotCT string
	var gotBody resendPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"abc"}`))
	}))
	defer srv.Close()

	s := NewResendSender("re_key", "H1 <noreply@example.com>").WithBaseURL(srv.URL)
	if err := s.Send(context.Background(), "ana@example.com", "Asunto", "<p>hola</p>", "hola"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost || gotPath != "/emails" {
		t.Errorf("got %s %s, want POST /emails", gotMethod, gotPath)
	}
	if gotAuth != "Bearer re_key" {
		t.Errorf("got Authorization %q", gotAuth)
	}
	if gotCT != "application/json" {
		t.Errorf("got Content-Type %q", gotCT)
	}
	if gotBody.From != "H1 <noreply@example.com>" || len(gotBody.To) != 1 || gotBody.To[0] != "ana@example.com" ||
		gotBody.Subject != "Asunto" || gotBody.HTML != "<p>hola</p>" || gotBody.Text != "hola" {
		t.Errorf("unexpected payload: %+v", gotBody)
	}
}

func TestResendSender_Send_IdempotencyKeyHeader(t *testing.T) {
	var gotKey string
	var present bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Idempotency-Key")
		_, present = r.Header["Idempotency-Key"]
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	s := NewResendSender("k", "f").WithBaseURL(srv.URL)

	ctx := WithIdempotencyKey(context.Background(), "booking-confirmation-7")
	if err := s.Send(ctx, "a@b.c", "s", "h", "t"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotKey != "booking-confirmation-7" {
		t.Errorf("got Idempotency-Key %q, want booking-confirmation-7", gotKey)
	}

	if err := s.Send(context.Background(), "a@b.c", "s", "h", "t"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if present {
		t.Errorf("Idempotency-Key header should be absent when not set, got %q", gotKey)
	}
}

func TestResendSender_Send_Non2xxReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"invalid from"}`))
	}))
	defer srv.Close()

	s := NewResendSender("k", "f").WithBaseURL(srv.URL)
	err := s.Send(context.Background(), "a@b.c", "s", "h", "t")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "422") || !strings.Contains(err.Error(), "invalid from") {
		t.Errorf("error should include status and body, got %q", err)
	}
}

func TestResendSender_Send_TruncatesLongErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(strings.Repeat("x", 5000)))
	}))
	defer srv.Close()

	s := NewResendSender("k", "f").WithBaseURL(srv.URL)
	err := s.Send(context.Background(), "a@b.c", "s", "h", "t")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(err.Error()) > 500 {
		t.Errorf("error not truncated, len=%d", len(err.Error()))
	}
}

func TestLogSender_Send_ReturnsNil(t *testing.T) {
	if err := (LogSender{}).Send(context.Background(), "a@b.c", "s", "<p>secret</p>", "secret"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
