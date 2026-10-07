// Package email contains transactional email senders.
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	defaultResendBaseURL = "https://api.resend.com"
	maxErrorBodyBytes    = 300
)

type idempotencyKeyCtx struct{}

// WithIdempotencyKey returns a context carrying a key that ResendSender sends
// as the Idempotency-Key header, so retries of the same logical email are
// deduplicated by Resend. Other senders ignore it.
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKeyCtx{}, key)
}

// IdempotencyKeyFromContext returns the key set by WithIdempotencyKey, or "".
func IdempotencyKeyFromContext(ctx context.Context) string {
	return idempotencyKeyFrom(ctx)
}

func idempotencyKeyFrom(ctx context.Context) string {
	key, _ := ctx.Value(idempotencyKeyCtx{}).(string)
	return key
}

// ResendSender sends email through the Resend HTTP API.
type ResendSender struct {
	apiKey  string
	from    string
	baseURL string
	client  *http.Client
}

func NewResendSender(apiKey, from string) *ResendSender {
	return &ResendSender{
		apiKey:  apiKey,
		from:    from,
		baseURL: defaultResendBaseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// WithBaseURL overrides the API base URL (used in tests).
func (s *ResendSender) WithBaseURL(baseURL string) *ResendSender {
	s.baseURL = baseURL
	return s
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
}

// Send delivers one email. Any non-2xx response is returned as an error.
func (s *ResendSender) Send(ctx context.Context, to, subject, html, text string) error {
	body, err := json.Marshal(resendPayload{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
		Text:    text,
	})
	if err != nil {
		return fmt.Errorf("resend: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("resend: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if key := idempotencyKeyFrom(ctx); key != "" {
		req.Header.Set("Idempotency-Key", key)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
		return fmt.Errorf("resend: unexpected status %d: %s", resp.StatusCode, snippet)
	}
	return nil
}

// LogSender is a development fallback that only logs recipient and subject.
// It never logs the body.
type LogSender struct{}

func (LogSender) Send(_ context.Context, to, subject, _, _ string) error {
	log.Printf("email (log only): to=%s subject=%q", to, subject)
	return nil
}
