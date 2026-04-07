package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"childcare-backend/middleware"

	"github.com/gin-gonic/gin"
)

func setupRequireFamilyRouter(secret string) *gin.Engine {
	r := gin.New()
	r.GET("/family-only",
		middleware.AuthMiddleware(secret),
		middleware.RequireFamily(),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		},
	)
	return r
}

func TestRequireFamily_WithFamily(t *testing.T) {
	r := setupRequireFamilyRouter(testSecret)
	token := makeToken("user1", "family-abc", "owner", testSecret, time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/family-only", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequireFamily_WithoutFamily(t *testing.T) {
	r := setupRequireFamilyRouter(testSecret)
	// empty familyID in token → user has no family
	token := makeToken("user2", "", "", testSecret, time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/family-only", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}
