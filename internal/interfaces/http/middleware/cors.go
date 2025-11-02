package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig represents CORS configuration
type CORSConfig struct {
	// AllowedOrigins is a list of origins allowed to make requests
	AllowedOrigins []string `json:"allowed_origins"`
	
	// AllowedMethods is a list of HTTP methods allowed
	AllowedMethods []string `json:"allowed_methods"`
	
	// AllowedHeaders is a list of headers allowed
	AllowedHeaders []string `json:"allowed_headers"`
	
	// ExposedHeaders is a list of headers exposed to the client
	ExposedHeaders []string `json:"exposed_headers"`
	
	// AllowCredentials indicates whether requests can include credentials
	AllowCredentials bool `json:"allow_credentials"`
	
	// MaxAge indicates how long the results of a preflight request can be cached
	MaxAge int `json:"max_age"`
	
	// Debug enables debug logging
	Debug bool `json:"debug"`
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"X-Request-ID",
		},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
		Debug:            false,
	}
}

// CORS returns a CORS middleware with the given configuration
func CORS(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	// Normalize origins
	normalizedOrigins := make([]string, len(config.AllowedOrigins))
	for i, origin := range config.AllowedOrigins {
		normalizedOrigins[i] = strings.ToLower(origin)
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Set CORS headers
		if origin != "" && isOriginAllowed(origin, normalizedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(normalizedOrigins) == 1 && normalizedOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		
		if len(config.ExposedHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Max-Age", string(rune(config.MaxAge)))

		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// isOriginAllowed checks if the origin is allowed
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// Support wildcard subdomains
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, domain) {
				originParts := strings.Split(origin, ".")
				if len(originParts) >= 2 {
					return true
				}
			}
		}
	}
	return false
}

// CORSWithOrigins returns a CORS middleware with specific allowed origins
func CORSWithOrigins(origins ...string) gin.HandlerFunc {
	config := DefaultCORSConfig()
	config.AllowedOrigins = origins
	return CORS(config)
}

// CORSWithCredentials returns a CORS middleware with credentials allowed
func CORSWithCredentials() gin.HandlerFunc {
	config := DefaultCORSConfig()
	config.AllowCredentials = true
	config.AllowedOrigins = []string{} // Must specify origins when credentials are enabled
	return CORS(config)
}

// RestrictiveCORS returns a restrictive CORS configuration for production
func RestrictiveCORS(allowedOrigins []string) gin.HandlerFunc {
	config := &CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
		Debug:            false,
	}
	return CORS(config)
}