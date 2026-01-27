package middleware

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var nonDigits = regexp.MustCompile(`\D`)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ts := time.Now().UnixNano()
		id := nonDigits.ReplaceAllString(time.Now().Format("150405.000000000"), "") + "-" + strings.TrimSpace(strings.ReplaceAll(strings.TrimLeft(strings.TrimPrefix(c.ClientIP(), "::ffff:"), ":"), ":", ""))
		c.Writer.Header().Set("x-request-id", id)
		c.Writer.Header().Set("x-request-ts", fmt.Sprintf("%d", ts))
		c.Next()
	}
}
