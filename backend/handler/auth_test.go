package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"childcare-backend/config"
	"childcare-backend/handler"
	"childcare-backend/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ─── Mock stores ─────────────────────────────────────────────────────────────

type mockUserStore struct {
	users map[string]*model.User // keyed by ID
}

func newMockUserStore() *mockUserStore {
	return &mockUserStore{users: make(map[string]*model.User)}
}

func (m *mockUserStore) Create(u *model.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserStore) GetByUsername(username string) (*model.User, error) {
	for _, u := range m.users {
		if u.Username != nil && *u.Username == username {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockUserStore) GetByOpenID(openid string) (*model.User, error) {
	for _, u := range m.users {
		if u.WxOpenID != nil && *u.WxOpenID == openid {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockUserStore) GetByID(id string) (*model.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *mockUserStore) UpdateFamily(userID, familyID, role string) error {
	u, ok := m.users[userID]
	if !ok {
		return errors.New("not found")
	}
	u.FamilyID = &familyID
	u.Role = &role
	return nil
}

// ─── Mock family store ───────────────────────────────────────────────────────

type mockFamilyStore struct {
	families map[string]*model.Family
}

func newMockFamilyStore() *mockFamilyStore {
	return &mockFamilyStore{families: make(map[string]*model.Family)}
}

func (m *mockFamilyStore) Create(f *model.Family) error {
	m.families[f.ID] = f
	return nil
}

func (m *mockFamilyStore) GetByID(id string) (*model.Family, error) {
	f, ok := m.families[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return f, nil
}

func (m *mockFamilyStore) GetMembers(familyID string) ([]*model.User, error) {
	return nil, nil
}

// ─── Mock WxClient ───────────────────────────────────────────────────────────

type mockWxClient struct {
	openID string
	err    error
}

func (m *mockWxClient) GetOpenID(appID, secret, code string) (string, error) {
	return m.openID, m.err
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret:        "test-jwt-secret",
		JWTRefreshSecret: "test-refresh-secret",
		WXAppID:          "wx-app-id",
		WXSecret:         "wx-secret",
		Port:             "8080",
	}
}

func newTestRouter(h *handler.AuthHandler) *gin.Engine {
	r := gin.New()
	r.POST("/api/auth/register", h.Register)
	r.POST("/api/auth/login", h.Login)
	r.POST("/api/auth/wx-login", h.WxLogin)
	r.POST("/api/auth/refresh", h.Refresh)
	return r
}

func postJSON(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// ─── Register tests ──────────────────────────────────────────────────────────

func TestRegister_HappyPath(t *testing.T) {
	us := newMockUserStore()
	fs := newMockFamilyStore()
	h := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), nil)
	r := newTestRouter(h)

	w := postJSON(r, "/api/auth/register", map[string]string{
		"username":    "alice",
		"password":    "pass123",
		"family_name": "Smith Family",
		"nickname":    "Alice",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if _, ok := resp["token"]; !ok {
		t.Fatalf("token missing in response: %s", w.Body.String())
	}
	if _, ok := resp["refresh_token"]; !ok {
		t.Fatalf("refresh_token missing in response: %s", w.Body.String())
	}
	user, ok := resp["user"].(map[string]any)
	if !ok {
		t.Fatalf("user object missing in response: %s", w.Body.String())
	}
	if user["id"] == "" {
		t.Fatalf("user.id empty: %v", user)
	}
}

func TestRegister_UsernameTaken(t *testing.T) {
	us := newMockUserStore()
	fs := newMockFamilyStore()

	// Pre-seed a user with username "alice".
	username := "alice"
	us.users["existing-id"] = &model.User{ID: "existing-id", Username: &username, Nickname: "Alice"}

	h := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), nil)
	r := newTestRouter(h)

	w := postJSON(r, "/api/auth/register", map[string]string{
		"username":    "alice",
		"password":    "pass123",
		"family_name": "Smith Family",
		"nickname":    "Alice",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "USERNAME_TAKEN" {
		t.Fatalf("expected USERNAME_TAKEN, got: %v", resp["code"])
	}
}

func TestRegister_InvalidRequest(t *testing.T) {
	us := newMockUserStore()
	fs := newMockFamilyStore()
	h := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), nil)
	r := newTestRouter(h)

	// password too short
	w := postJSON(r, "/api/auth/register", map[string]string{
		"username":    "alice",
		"password":    "ab",
		"family_name": "Smith Family",
		"nickname":    "Alice",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ─── Login tests ─────────────────────────────────────────────────────────────

func setupUserWithPassword(us *mockUserStore, username, password, nickname string) *model.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	h := string(hash)
	u := &model.User{
		ID:           "user-id-1",
		Username:     &username,
		PasswordHash: &h,
		Nickname:     nickname,
	}
	us.users[u.ID] = u
	return u
}

func TestLogin_ValidCredentials(t *testing.T) {
	us := newMockUserStore()
	fs := newMockFamilyStore()
	setupUserWithPassword(us, "alice", "pass123", "Alice")

	h := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), nil)
	r := newTestRouter(h)

	w := postJSON(r, "/api/auth/login", map[string]string{
		"username": "alice",
		"password": "pass123",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["token"]; !ok {
		t.Fatal("token missing")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	us := newMockUserStore()
	fs := newMockFamilyStore()
	setupUserWithPassword(us, "alice", "pass123", "Alice")

	h := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), nil)
	r := newTestRouter(h)

	w := postJSON(r, "/api/auth/login", map[string]string{
		"username": "alice",
		"password": "wrongpass",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	us := newMockUserStore()
	fs := newMockFamilyStore()

	h := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), nil)
	r := newTestRouter(h)

	w := postJSON(r, "/api/auth/login", map[string]string{
		"username": "nobody",
		"password": "pass123",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

// ─── WxLogin tests ───────────────────────────────────────────────────────────

func TestWxLogin_NewUser(t *testing.T) {
	us := newMockUserStore()
	fs := newMockFamilyStore()
	wxClient := &mockWxClient{openID: "wx-open-id-123"}

	h := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), wxClient)
	r := newTestRouter(h)

	w := postJSON(r, "/api/auth/wx-login", map[string]string{
		"code": "wx-code-abc",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["token"]; !ok {
		t.Fatal("token missing")
	}
	// The new user should have been stored.
	found, err := us.GetByOpenID("wx-open-id-123")
	if err != nil || found == nil {
		t.Fatal("new wx user not created in store")
	}
}

func TestWxLogin_ExistingUser(t *testing.T) {
	us := newMockUserStore()
	fs := newMockFamilyStore()

	openID := "existing-openid"
	existing := &model.User{
		ID:       "existing-wx-user",
		WxOpenID: &openID,
		Nickname: "微信用户",
	}
	us.users[existing.ID] = existing

	wxClient := &mockWxClient{openID: openID}
	h := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), wxClient)
	r := newTestRouter(h)

	w := postJSON(r, "/api/auth/wx-login", map[string]string{
		"code": "any-code",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestWxLogin_WxClientError(t *testing.T) {
	us := newMockUserStore()
	fs := newMockFamilyStore()
	wxClient := &mockWxClient{err: errors.New("wx api error")}

	h := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), wxClient)
	r := newTestRouter(h)

	w := postJSON(r, "/api/auth/wx-login", map[string]string{
		"code": "bad-code",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ─── Refresh tests ────────────────────────────────────────────────────────────

func getRefreshToken(t *testing.T, r *gin.Engine) string {
	t.Helper()
	us := newMockUserStore()
	fs := newMockFamilyStore()
	setupUserWithPassword(us, "alice", "pass123", "Alice")
	h2 := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), nil)
	r2 := newTestRouter(h2)

	w := postJSON(r2, "/api/auth/login", map[string]string{
		"username": "alice",
		"password": "pass123",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp["refresh_token"].(string)
}

func TestRefresh_ValidToken(t *testing.T) {
	us := newMockUserStore()
	fs := newMockFamilyStore()
	setupUserWithPassword(us, "alice", "pass123", "Alice")

	h := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), nil)
	r := newTestRouter(h)

	// Get a real refresh token by logging in.
	wLogin := postJSON(r, "/api/auth/login", map[string]string{
		"username": "alice",
		"password": "pass123",
	})
	if wLogin.Code != http.StatusOK {
		t.Fatalf("login failed: %d", wLogin.Code)
	}
	var loginResp map[string]any
	json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	refreshToken := loginResp["refresh_token"].(string)

	// Now refresh.
	w := postJSON(r, "/api/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["token"]; !ok {
		t.Fatal("new access token missing")
	}
	if _, ok := resp["refresh_token"]; !ok {
		t.Fatal("new refresh token missing")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	us := newMockUserStore()
	fs := newMockFamilyStore()
	h := handler.NewAuthHandlerWithWxClient(us, fs, testConfig(), nil)
	r := newTestRouter(h)

	w := postJSON(r, "/api/auth/refresh", map[string]string{
		"refresh_token": "this.is.not.valid",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
