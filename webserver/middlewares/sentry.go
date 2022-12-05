package middlewares

import (
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func Sentry() gin.HandlerFunc {
	return sentrygin.New(sentrygin.Options{})
}
