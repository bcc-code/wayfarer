package members

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrMemberNotFound means the person is definitively gone, not a transient failure
var ErrMemberNotFound = errors.New("member not found")

// newStatusError wraps ErrMemberNotFound on 404 so callers can errors.Is against it
func newStatusError(statusCode int, body []byte) error {
	if statusCode == http.StatusNotFound {
		return fmt.Errorf("members API returned status %d: %s: %w", statusCode, string(body), ErrMemberNotFound)
	}
	return fmt.Errorf("members API returned status %d: %s", statusCode, string(body))
}

// IsBreakerSuccess treats 404s as non-failures so a run of deleted members can't trip the breaker
func IsBreakerSuccess(err error) bool {
	return err == nil || errors.Is(err, ErrMemberNotFound)
}
