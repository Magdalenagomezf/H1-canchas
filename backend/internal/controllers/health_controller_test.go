package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakePinger struct{ err error }

func (f fakePinger) PingContext(context.Context) error { return f.err }

func doHealth(t *testing.T, p dbPinger) (int, map[string]string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", NewHealthController(p, "abc123").Health)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body %q: %v", w.Body.String(), err)
	}
	return w.Code, body
}

func TestHealth_DBReachable_ReturnsOKWithCommit(t *testing.T) {
	code, body := doHealth(t, fakePinger{})

	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	want := map[string]string{"status": "ok", "db": "connected", "commit": "abc123"}
	for k, v := range want {
		if body[k] != v {
			t.Errorf("%s = %q, want %q", k, body[k], v)
		}
	}
}

func TestHealth_DBUnreachable_Returns503WithCommit(t *testing.T) {
	code, body := doHealth(t, fakePinger{err: errors.New("connection refused")})

	if code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", code)
	}
	want := map[string]string{"status": "degraded", "db": "unreachable", "commit": "abc123"}
	for k, v := range want {
		if body[k] != v {
			t.Errorf("%s = %q, want %q", k, body[k], v)
		}
	}
}
