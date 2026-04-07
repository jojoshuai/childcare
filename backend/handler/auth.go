package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"childcare-backend/config"
	"childcare-backend/model"
	"childcare-backend/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ─── WxClient ────────────────────────────────────────────────────────────────

// WxClient abstracts the WeChat code-exchange call so it can be mocked in tests.
type WxClient interface {
	GetOpenID(appID, secret, code string) (openid string, err error)
}

// realWxClient calls the actual WeChat API.
type realWxClient struct{}

type wxSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func (r *realWxClient) GetOpenID(appID, secret, code string) (string, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, secret, code,
	)
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("wx request failed: %w", err)
	}
	defer resp.Body.Close()

	var result wxSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("wx response decode failed: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wx error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result.OpenID, nil
}

// ─── AuthHandler ─────────────────────────────────────────────────────────────

// AuthHandler holds the dependencies needed by the auth HTTP handlers.
type AuthHandler struct {
	userStore   store.UserStore
	familyStore store.FamilyStore
	cfg         *config.Config
	wxClient    WxClient
}

// NewAuthHandler constructs an AuthHandler with the real WxClient.
func NewAuthHandler(userStore store.UserStore, familyStore store.FamilyStore, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userStore:   userStore,
		familyStore: familyStore,
		cfg:         cfg,
		wxClient:    &realWxClient{},
	}
}

// NewAuthHandlerWithWxClient constructs an AuthHandler with a custom WxClient.
// Pass nil to use the real WeChat client. Intended for testing.
func NewAuthHandlerWithWxClient(userStore store.UserStore, familyStore store.FamilyStore, cfg *config.Config, wxc WxClient) *AuthHandler {
	if wxc == nil {
		wxc = &realWxClient{}
	}
	return &AuthHandler{
		userStore:   userStore,
		familyStore: familyStore,
		cfg:         cfg,
		wxClient:    wxc,
	}
}

// ─── Claims ──────────────────────────────────────────────────────────────────

type authClaims struct {
	UserID   string `json:"user_id"`
	FamilyID string `json:"family_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// generateTokens mints a new access token (7 days) and refresh token (30 days).
func generateTokens(userID, familyID, role, accessSecret, refreshSecret string) (accessToken, refreshToken string, err error) {
	now := time.Now()

	makeClaims := func(exp time.Duration) authClaims {
		return authClaims{
			UserID:   userID,
			FamilyID: familyID,
			Role:     role,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userID,
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(exp)),
			},
		}
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, makeClaims(7*24*time.Hour)).SignedString([]byte(accessSecret))
	if err != nil {
		return "", "", fmt.Errorf("sign access token: %w", err)
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, makeClaims(30*24*time.Hour)).SignedString([]byte(refreshSecret))
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token: %w", err)
	}
	return accessToken, refreshToken, nil
}

// userResponse is the "user" sub-object returned in auth responses.
type userResponse struct {
	ID       string  `json:"id"`
	Nickname string  `json:"nickname"`
	FamilyID *string `json:"family_id"`
	Role     *string `json:"role"`
}

func userToResponse(u *model.User) userResponse {
	return userResponse{
		ID:       u.ID,
		Nickname: u.Nickname,
		FamilyID: u.FamilyID,
		Role:     u.Role,
	}
}

// familyIDStr returns empty string when the pointer is nil, otherwise the value.
func familyIDStr(u *model.User) string {
	if u.FamilyID == nil {
		return ""
	}
	return *u.FamilyID
}

// roleStr returns empty string when the pointer is nil, otherwise the value.
func roleStr(u *model.User) string {
	if u.Role == nil {
		return ""
	}
	return *u.Role
}

// ─── Register ────────────────────────────────────────────────────────────────

type registerRequest struct {
	Username   string `json:"username"   binding:"required,min=3,max=50"`
	Password   string `json:"password"   binding:"required,min=6,max=100"`
	FamilyName string `json:"family_name" binding:"required,min=1,max=100"`
	Nickname   string `json:"nickname"   binding:"required,min=1,max=50"`
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// 1. Check username uniqueness.
	existing, err := h.userStore.GetByUsername(req.Username)
	if err == nil && existing != nil {
		errorResponse(c, http.StatusBadRequest, "USERNAME_TAKEN", "该用户名已被占用")
		return
	}

	// 2. Hash password.
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "password hashing failed")
		return
	}
	passwordHash := string(hashBytes)

	// 3. Create family.
	family := &model.Family{
		ID:   uuid.NewString(),
		Name: req.FamilyName,
	}
	if err := h.familyStore.Create(family); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create family")
		return
	}

	// 4. Create user.
	role := "owner"
	user := &model.User{
		ID:           uuid.NewString(),
		Username:     &req.Username,
		PasswordHash: &passwordHash,
		Nickname:     req.Nickname,
		FamilyID:     &family.ID,
		Role:         &role,
	}
	if err := h.userStore.Create(user); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user")
		return
	}

	// 5. Generate tokens.
	accessToken, refreshToken, err := generateTokens(user.ID, family.ID, role, h.cfg.JWTSecret, h.cfg.JWTRefreshSecret)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "token generation failed")
		return
	}

	// 6. Respond.
	c.JSON(http.StatusCreated, gin.H{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"user":          userToResponse(user),
	})
}

// ─── Login ───────────────────────────────────────────────────────────────────

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// 1. Find user.
	user, err := h.userStore.GetByUsername(req.Username)
	if err != nil || user == nil {
		errorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误")
		return
	}

	// 2. Verify password.
	if user.PasswordHash == nil {
		errorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		errorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误")
		return
	}

	// 3. Generate tokens.
	accessToken, refreshToken, err := generateTokens(user.ID, familyIDStr(user), roleStr(user), h.cfg.JWTSecret, h.cfg.JWTRefreshSecret)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "token generation failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"user":          userToResponse(user),
	})
}

// ─── WxLogin ─────────────────────────────────────────────────────────────────

type wxLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// WxLogin handles POST /api/auth/wx-login.
func (h *AuthHandler) WxLogin(c *gin.Context) {
	var req wxLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// 1. Exchange code for openid.
	openid, err := h.wxClient.GetOpenID(h.cfg.WXAppID, h.cfg.WXSecret, req.Code)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "WX_ERROR", err.Error())
		return
	}

	// 2. Find or create user.
	user, err := h.userStore.GetByOpenID(openid)
	if err != nil || user == nil {
		// 3. Create new wx user (no family, no role).
		nickname := "微信用户"
		user = &model.User{
			ID:       uuid.NewString(),
			WxOpenID: &openid,
			Nickname: nickname,
		}
		if createErr := h.userStore.Create(user); createErr != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user")
			return
		}
	}

	// 4. Generate tokens.
	accessToken, refreshToken, err := generateTokens(user.ID, familyIDStr(user), roleStr(user), h.cfg.JWTSecret, h.cfg.JWTRefreshSecret)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "token generation failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"user":          userToResponse(user),
	})
}

// ─── Refresh ─────────────────────────────────────────────────────────────────

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh handles POST /api/auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// 1. Parse and validate the refresh token.
	claims := &authClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(h.cfg.JWTRefreshSecret), nil
	})
	if err != nil || !token.Valid {
		msg := "invalid or expired refresh token"
		if err != nil {
			msg = err.Error()
		}
		errorResponse(c, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", msg)
		return
	}

	// 2. Verify user still exists.
	user, err := h.userStore.GetByID(claims.UserID)
	if err != nil || user == nil {
		errorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}

	// 3. Issue new token pair.
	accessToken, refreshToken, err := generateTokens(user.ID, familyIDStr(user), roleStr(user), h.cfg.JWTSecret, h.cfg.JWTRefreshSecret)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "token generation failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}
