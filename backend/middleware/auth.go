package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the JWT payload for access and refresh tokens.
type Claims struct {
	UserID   string `json:"user_id"`
	FamilyID string `json:"family_id"` // empty string if null
	Role     string `json:"role"`       // empty string if null
	jwt.RegisteredClaims
}

// AuthMiddleware returns a gin.HandlerFunc that validates the JWT access token.
// It reads the Authorization: Bearer <token> header. On success it sets
// "userID", "familyID", and "role" in the Gin context. On failure it returns
// 401 and aborts.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "missing Authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "invalid Authorization header format",
			})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			msg := "invalid or expired token"
			if err != nil {
				msg = err.Error()
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": msg,
			})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("familyID", claims.FamilyID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// GetUserID retrieves the userID from the Gin context (set by AuthMiddleware).
func GetUserID(c *gin.Context) string { return c.GetString("userID") }

// GetFamilyID retrieves the familyID from the Gin context (set by AuthMiddleware).
func GetFamilyID(c *gin.Context) string { return c.GetString("familyID") }

// GetRole retrieves the role from the Gin context (set by AuthMiddleware).
func GetRole(c *gin.Context) string { return c.GetString("role") }
