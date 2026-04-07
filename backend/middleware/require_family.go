package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireFamily must be used after AuthMiddleware. It returns 403 if the
// familyID stored in the context is an empty string (i.e. the user has not
// joined a family yet).
func RequireFamily() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetFamilyID(c) == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "NO_FAMILY_JOINED",
				"message": "请先通过邀请码加入家庭",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
