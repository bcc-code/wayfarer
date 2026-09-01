package autotls

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Validation(t *testing.T) {
	_, err := New(nil, t.TempDir(), "")
	assert.Error(t, err, "no domains")

	_, err = New([]string{"a.example.com", ""}, t.TempDir(), "")
	assert.Error(t, err, "empty domain")

	_, err = New([]string{"a.example.com"}, "", "")
	assert.Error(t, err, "no cache dir")
}

func TestNew_CreatesCacheDirAndHostPolicy(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "autocert")
	m, err := New([]string{"interact.bcc.no", "interact.bcc.media"}, dir, "ops@example.com")
	require.NoError(t, err)
	assert.DirExists(t, dir)

	assert.True(t, m.AllowsHost("interact.bcc.no"))
	assert.True(t, m.AllowsHost("interact.bcc.media"))
	assert.False(t, m.AllowsHost("evil.example.com"))

	cfg := m.TLSConfig()
	require.NotNil(t, cfg)
	// tls-alpn-01 must be negotiable on the listener.
	assert.Contains(t, cfg.NextProtos, "acme-tls/1")
}
