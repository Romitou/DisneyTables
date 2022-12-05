package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if strings.TrimPrefix(authorization, "Bearer ") != os.Getenv("WEBSERVER_TOKEN") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}
