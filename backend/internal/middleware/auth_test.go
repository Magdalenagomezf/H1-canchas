package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtutil "H1-canchas/pkg"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testSecret = "test-secret"

func newAuthRouter(secret string, next gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.GET("/protected", Auth(secret), next)
	return r
}

func okHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user_id": c.GetInt64("user_id"),
		"role":    c.GetString("role"),
	})
}

// --- Auth middleware ---

func TestAuth_NoHeader(t *testing.T) {
	r := newAuthRouter(testSecret, okHandler)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestAuth_MissingBearerPrefix(t *testing.T) {
	r := newAuthRouter(testSecret, okHandler)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "token-without-bearer")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	r := newAuthRouter(testSecret, okHandler)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestAuth_WrongSecret(t *testing.T) {
	token, _ := jwtutil.GenerateToken(1, "customer", "other-secret")
	r := newAuthRouter(testSecret, okHandler)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	claims := &jwtutil.Claims{
		UserID: 1,
		Role:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(testSecret))

	r := newAuthRouter(testSecret, okHandler)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestAuth_ValidToken_SetsContext(t *testing.T) {
	const userID int64 = 42
	const role = "receptionist"
	token, _ := jwtutil.GenerateToken(userID, role, testSecret)

	var gotUserID int64
	var gotRole string
	handler := func(c *gin.Context) {
		gotUserID = c.GetInt64("user_id")
		gotRole = c.GetString("role")
		c.Status(http.StatusOK)
	}

	r := newAuthRouter(testSecret, handler)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
	if gotUserID != userID {
		t.Errorf("got user_id %d, want %d", gotUserID, userID)
	}
	if gotRole != role {
		t.Errorf("got role %q, want %q", gotRole, role)
	}
}

// --- RequireRole middleware ---

func newRoleRouter(role string, requiredRoles ...string) *gin.Engine {
	r := gin.New()
	r.GET("/action", func(c *gin.Context) {
		c.Set("role", role)
		c.Next()
	}, RequireRole(requiredRoles...), okHandler)
	return r
}

func TestRequireRole_MatchingRole(t *testing.T) {
	r := newRoleRouter("admin", "admin", "receptionist")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/action", nil))
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
}

func TestRequireRole_NonMatchingRole(t *testing.T) {
	r := newRoleRouter("customer", "admin", "receptionist")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/action", nil))
	if w.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", w.Code)
	}
}

func TestRequireRole_SingleRoleMatch(t *testing.T) {
	r := newRoleRouter("receptionist", "receptionist")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/action", nil))
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
}
