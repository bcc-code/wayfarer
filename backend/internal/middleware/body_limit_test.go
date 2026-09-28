package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMaxBodyBytes(t *testing.T) {
	router := gin.New()
	router.POST("/", MaxBodyBytes(16), func(c *gin.Context) {
		if _, err := io.ReadAll(c.Request.Body); err != nil {
			c.Status(http.StatusRequestEntityTooLarge)
			return
		}
		c.Status(http.StatusOK)
	})

	send := func(body string, chunked bool) int {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		if chunked {
			req.ContentLength = -1
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}

	assert.Equal(t, http.StatusOK, send("small", false))
	assert.Equal(t, http.StatusOK, send(strings.Repeat("a", 16), false))
	assert.Equal(t, http.StatusRequestEntityTooLarge, send(strings.Repeat("a", 17), false), "declared length over the limit")
	assert.Equal(t, http.StatusRequestEntityTooLarge, send(strings.Repeat("a", 1000), true), "undeclared length over the limit")
}
