package middleware

import "github.com/gin-gonic/gin"

// Content-Security-Policy for the SPA + API (same origin). Allows:
//   - scripts only from self (no inline script),
//   - inline styles (the app applies per-form themes via style={} attributes) and
//     Google Fonts stylesheets,
//   - fonts from Google's CDN, images from self + data: URIs,
//   - XHR/fetch to self; framing denied (clickjacking), base/forms locked to self.
const contentSecurityPolicy = "default-src 'self'; " +
	"base-uri 'self'; " +
	"object-src 'none'; " +
	"frame-ancestors 'none'; " +
	"img-src 'self' data:; " +
	"script-src 'self'; " +
	"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
	"font-src 'self' https://fonts.gstatic.com; " +
	"connect-src 'self'; " +
	"form-action 'self'"

// SecurityHeaders sets defensive response headers on every request. hsts adds
// Strict-Transport-Security (only meaningful — and only set — when served over
// HTTPS in production); browsers ignore HSTS on plain HTTP.
func SecurityHeaders(hsts bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		if hsts {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
