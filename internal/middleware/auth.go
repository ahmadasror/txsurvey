package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/pkg/auth"
	"github.com/ahmadasror/txsurvey/pkg/response"
)

// SessionCookieName is the httpOnly cookie holding the app-minted session JWT.
const SessionCookieName = "session"

// SessionAuth gates creator-only routes. It reads the session cookie (NOT an
// Authorization header — this app is its own IdP), validates it, and stores the
// creator identity in the context for handlers.
func SessionAuth(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := c.Cookie(SessionCookieName)
		if err != nil || raw == "" {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			c.Abort()
			return
		}
		claims, err := jwtMgr.ValidateSessionToken(raw)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired session")
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("name", claims.Name)
		c.Next()
	}
}
