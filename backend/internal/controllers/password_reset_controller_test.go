package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"H1-canchas/internal/service"

	"github.com/gin-gonic/gin"
)

type mockPasswordResetService struct {
	requested []string
	confirmFn func(ctx context.Context, token, password string) error
	checkFn   func(ctx context.Context, token string) error
}

func (m *mockPasswordResetService) CheckToken(ctx context.Context, token string) error {
	if m.checkFn != nil {
		return m.checkFn(ctx, token)
	}
	return nil
}

func (m *mockPasswordResetService) RequestReset(email string) {
	m.requested = append(m.requested, email)
}

func (m *mockPasswordResetService) Confirm(ctx context.Context, token, password string) error {
	if m.confirmFn != nil {
		return m.confirmFn(ctx, token, password)
	}
	return nil
}

func newPasswordResetRouter(svc passwordResetServiceI) *gin.Engine {
	r := gin.New()
	ctrl := NewPasswordResetController(svc)
	r.POST("/auth/password-reset", ctrl.Request)
	r.POST("/auth/password-reset/confirm", ctrl.Confirm)
	r.POST("/auth/password-reset/check", ctrl.Check)
	return r
}

func postJSON(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestPasswordResetRequest_Returns202AndStartsReset(t *testing.T) {
	svc := &mockPasswordResetService{}
	w := postJSON(newPasswordResetRouter(svc), "/auth/password-reset", `{"email":"ana@example.com"}`)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", w.Code)
	}
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != passwordResetRequestedMessage {
		t.Errorf("message = %q, want %q", resp["message"], passwordResetRequestedMessage)
	}
	if len(svc.requested) != 1 || svc.requested[0] != "ana@example.com" {
		t.Errorf("requested = %v, want [ana@example.com]", svc.requested)
	}
}

func TestPasswordResetRequest_MissingEmailIs400(t *testing.T) {
	svc := &mockPasswordResetService{}
	w := postJSON(newPasswordResetRouter(svc), "/auth/password-reset", `{}`)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if len(svc.requested) != 0 {
		t.Error("reset must not start without an email")
	}
}

func TestPasswordResetConfirm_Success204(t *testing.T) {
	var gotToken, gotPassword string
	svc := &mockPasswordResetService{confirmFn: func(_ context.Context, token, password string) error {
		gotToken, gotPassword = token, password
		return nil
	}}
	w := postJSON(newPasswordResetRouter(svc), "/auth/password-reset/confirm", `{"token":"tok","password":"nueva-clave"}`)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
	if gotToken != "tok" || gotPassword != "nueva-clave" {
		t.Errorf("service got (%q, %q)", gotToken, gotPassword)
	}
}

func TestPasswordResetConfirm_ErrorMapping(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		status   int
		wantCode string
	}{
		{"invalid token", service.ErrInvalidResetToken, http.StatusBadRequest, "invalid_reset_token"},
		{"short password", service.ErrPasswordTooShort, http.StatusBadRequest, ""},
		{"long password", service.ErrPasswordTooLong, http.StatusBadRequest, ""},
		{"internal", errors.New("db down"), http.StatusInternalServerError, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockPasswordResetService{confirmFn: func(context.Context, string, string) error { return tc.err }}
			w := postJSON(newPasswordResetRouter(svc), "/auth/password-reset/confirm", `{"token":"t","password":"p"}`)

			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
			var resp map[string]string
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["code"] != tc.wantCode {
				t.Errorf("code = %q, want %q", resp["code"], tc.wantCode)
			}
			if tc.status == http.StatusInternalServerError && resp["error"] != "error interno" {
				t.Errorf("internal error leaked: %q", resp["error"])
			}
			if tc.status == http.StatusBadRequest && resp["error"] != tc.err.Error() {
				t.Errorf("error = %q, want %q", resp["error"], tc.err.Error())
			}
		})
	}
}

func TestPasswordResetCheck_StatusMapping(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		status   int
		wantCode string
	}{
		{"valid", nil, http.StatusNoContent, ""},
		{"invalid", service.ErrInvalidResetToken, http.StatusBadRequest, "invalid_reset_token"},
		{"internal", errors.New("db down"), http.StatusInternalServerError, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotToken string
			svc := &mockPasswordResetService{checkFn: func(_ context.Context, token string) error {
				gotToken = token
				return tc.err
			}}
			w := postJSON(newPasswordResetRouter(svc), "/auth/password-reset/check", `{"token":"tok"}`)

			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
			if gotToken != "tok" {
				t.Errorf("service got token %q, want tok", gotToken)
			}
			if tc.status == http.StatusNoContent {
				return
			}
			var resp map[string]string
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["code"] != tc.wantCode {
				t.Errorf("code = %q, want %q", resp["code"], tc.wantCode)
			}
			if tc.status == http.StatusInternalServerError && resp["error"] != "error interno" {
				t.Errorf("internal error leaked: %q", resp["error"])
			}
		})
	}
}
