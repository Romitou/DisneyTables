package middlewares

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("WEBSERVER_TOKEN", "secret-test-token")
	defer os.Unsetenv("WEBSERVER_TOKEN")

	r := gin.New()
	r.Use(Auth())
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Case 1: Missing Authorization Header
	reqNoAuth, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	wNoAuth := httptest.NewRecorder()
	r.ServeHTTP(wNoAuth, reqNoAuth)
	if wNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when header is missing, got %d", wNoAuth.Code)
	}

	// Case 2: Invalid Bearer Token
	reqInvalid, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	reqInvalid.Header.Set("Authorization", "Bearer wrong-token")
	wInvalid := httptest.NewRecorder()
	r.ServeHTTP(wInvalid, reqInvalid)
	if wInvalid.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for invalid token, got %d", wInvalid.Code)
	}

	// Case 3: Valid Bearer Token
	reqValid, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	reqValid.Header.Set("Authorization", "Bearer secret-test-token")
	wValid := httptest.NewRecorder()
	r.ServeHTTP(wValid, reqValid)
	if wValid.Code != http.StatusOK {
		t.Fatalf("expected status 200 for valid token, got %d", wValid.Code)
	}
}
