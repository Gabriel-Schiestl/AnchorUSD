package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimitMiddleware(limiter *rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
            if !limiter.Allow() {
                c.JSON(http.StatusTooManyRequests, map[string]string{
                    "error": "Too Many Requests",
                })
            }
            c.Next()
        }
}