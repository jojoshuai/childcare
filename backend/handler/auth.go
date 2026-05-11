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

type WxClient interface {
	GetOpenID(appID, secret, code string) (openid string, err error)
}

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
	resp, err := http.Get(url)
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

type AuthHandler struct {
	userStore store.UserStore
	cfg       *config.Config
	wxClient  WxClient
}

func NewAuthHandler(userStore store.UserStore, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userStore:  userStore,
		cfg:        cfg,
		wxClient:   &realWxClient{},
	}
}

type authClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func generateTokens(userID, secret string) (string, error) {
	now := time.Now()
	claims := authClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return token, nil
}

type userResponse struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
}

func userToResponse(u *model.User) userResponse {
	return userResponse{
		ID:       u.ID,
		Nickname: u.Nickname,
	}
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Nickname string `json:"nickname" binding:"required,min=1,max=50"`
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	existing, err := h.userStore.GetByUsername(req.Username)
	if err == nil && existing != nil {
		errorResponse(c, http.StatusBadRequest, "USERNAME_TAKEN", "该用户名已被占用")
		return
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "password hashing failed")
		return
	}

	hashStr := string(hashBytes)
	now := time.Now()
	user := &model.User{
		ID:           uuid.NewString(),
		Username:     &req.Username,
		PasswordHash: &hashStr,
		Nickname:     req.Nickname,
		CreatedAt:    now,
	}
	if err := h.userStore.Create(user); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user")
		return
	}

	token, err := generateTokens(user.ID, h.cfg.JWTSecret)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "token generation failed")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  userToResponse(user),
	})
}

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

	user, err := h.userStore.GetByUsername(req.Username)
	if err != nil || user == nil {
		errorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误")
		return
	}

	if user.PasswordHash == nil {
		errorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		errorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误")
		return
	}

	token, err := generateTokens(user.ID, h.cfg.JWTSecret)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "token generation failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  userToResponse(user),
	})
}

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

	openid, err := h.wxClient.GetOpenID(h.cfg.WXAppID, h.cfg.WXSecret, req.Code)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "WX_ERROR", err.Error())
		return
	}

	user, err := h.userStore.GetByOpenID(openid)
	if err != nil || user == nil {
		nickname := "微信用户"
		user = &model.User{
			ID:        uuid.NewString(),
			WxOpenID:  &openid,
			Nickname:  nickname,
			CreatedAt: time.Now(),
		}
		if createErr := h.userStore.Create(user); createErr != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user")
			return
		}
	}

	token, err := generateTokens(user.ID, h.cfg.JWTSecret)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "token generation failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  userToResponse(user),
	})
}

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

	claims := &authClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(t *jwt.Token) (any, error) {
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

	user, err := h.userStore.GetByID(claims.UserID)
	if err != nil || user == nil {
		errorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}

	accessToken, err := generateTokens(user.ID, h.cfg.JWTSecret)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "token generation failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": accessToken,
	})
}
