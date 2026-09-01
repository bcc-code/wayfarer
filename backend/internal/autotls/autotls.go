// Package autotls configures ACME (Let's Encrypt) certificate issuance and
// automatic renewal for the server's native TLS mode. Certificates are
// obtained via tls-alpn-01 on the TLS listener itself (the listener must be
// reachable from the internet on port 443 for issuance/renewal), cached on
// disk, and renewed automatically ~30 days before expiry.
package autotls

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/crypto/acme/autocert"
)

// Manager wraps autocert with the wiring the server needs.
type Manager struct {
	inner *autocert.Manager
}

// New builds a certificate manager for the given domains. cacheDir is
// created if missing; it must persist across restarts or every boot
// re-issues against Let's Encrypt rate limits. email is optional (expiry
// notices from the CA).
func New(domains []string, cacheDir, email string) (*Manager, error) {
	if len(domains) == 0 {
		return nil, fmt.Errorf("autotls: at least one domain is required")
	}
	for _, d := range domains {
		if d == "" {
			return nil, fmt.Errorf("autotls: empty domain in list")
		}
	}
	if cacheDir == "" {
		return nil, fmt.Errorf("autotls: cache dir is required")
	}
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return nil, fmt.Errorf("autotls: failed to create cache dir: %w", err)
	}

	return &Manager{
		inner: &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domains...),
			Cache:      autocert.DirCache(cacheDir),
			Email:      email,
		},
	}, nil
}

// TLSConfig returns the tls.Config for the HTTPS listener: it serves cached
// certificates, answers tls-alpn-01 challenges, and triggers issuance and
// renewal on demand.
func (m *Manager) TLSConfig() *tls.Config {
	return m.inner.TLSConfig()
}

// HTTPHandler returns a handler for an optional plain-HTTP listener: it
// answers http-01 challenges (fallback when tls-alpn-01 is unavailable) and
// redirects everything else to HTTPS.
func (m *Manager) HTTPHandler() http.Handler {
	return m.inner.HTTPHandler(nil)
}

// AllowsHost reports whether the manager would issue for host — exposed for
// tests and boot-time sanity logging.
func (m *Manager) AllowsHost(host string) bool {
	return m.inner.HostPolicy(nil, host) == nil
}
