package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"childcare-backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testSecret = "test-secret-key"

// makeToken creates a signed JWT using the given secret and expiry duration.
// A negative duration creates an already-expired token.
func makeToken(userID, familyID, role, secret string, exp time.Duration) string {
	claims := middleware.Claims{
		UserID:   userID,
		FamilyID: familyID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return token
}

func setupRouter(secret string) *gin.Engine {
	r := gin.New()
	r.GET("/protected", middleware.AuthMiddleware(secret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"userID":   middleware.GetUserID(c),
			"familyID": middleware.GetFamilyID(c),
			"role":     middleware.GetRole(c),
		})
	})
	return r
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	r := setupRouter(testSecret)
	token := makeToken("user1", "family1", "owner", testSecret, time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	r := setupRouter(testSecret)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	r := setupRouter(testSecret)
	token := makeToken("user1", "family1", "owner", testSecret, -time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	r := setupRouter(testSecret)
	token := makeToken("user1", "family1", "owner", "wrong-secret", time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	r := setupRouter(testSecret)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token somevalue")
	r.ServeHTTP(w, req)

	// "Token somevalue" passes format check but the token itself is invalid → 401
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_SetsContextValues(t *testing.T) {
	r := setupRouter(testSecret)
	token := makeToken("user42", "fam99", "member", testSecret, time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !contains(body, "user42") || !contains(body, "fam99") || !contains(body, "member") {
		t.Fatalf("context values not propagated: %s", body)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
