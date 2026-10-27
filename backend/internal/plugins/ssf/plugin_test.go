package ssf

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/plugins"
)

func TestPlugin_RegistersNewAndOldContentEventRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	err := NewPlugin(Config{}).Register(router, plugins.Dependencies{}, nil)
	require.NoError(t, err)

	routes := map[string]bool{}
	for _, r := range router.Routes() {
		routes[r.Method+" "+r.Path] = true
	}
	assert.True(t, routes[http.MethodPost+" /plugins/ssf/content-event"])
	assert.True(t, routes[http.MethodPost+" /plugins/ladder-to-heaven/content-event"],
		"the old path stays so existing webhook configs keep working")
}
