package middleware

import (
	"mocksmith/internal/server/store"

	"github.com/gin-gonic/gin"
)

func RequireAdmin(adminKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("x-admin-key") != adminKey {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireRuntimeKey(state *store.State) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("x-mocksmith-key")
		snap := state.Snapshot()
		if _, ok := snap.APIKeys[key]; !ok {
			c.JSON(401, gin.H{"error": "invalid api key"})
			c.Abort()
			return
		}
		if snap.RateLimit > 0 {
			lim := state.Limits().Get(key, snap.RateLimit)
			if !lim.Allow() {
				c.Header("Retry-After", "60")
				c.JSON(429, gin.H{"error": "rate limit"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
