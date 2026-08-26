package members

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStatusError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		body           []byte
		wantMemberGone bool
	}{
		{name: "404 wraps ErrMemberNotFound", statusCode: http.StatusNotFound, body: []byte(`{"error":{"code":"not-found","message":""}}`), wantMemberGone: true},
		{name: "500 does not wrap ErrMemberNotFound", statusCode: http.StatusInternalServerError, body: []byte("internal error"), wantMemberGone: false},
		{name: "403 does not wrap ErrMemberNotFound", statusCode: http.StatusForbidden, body: []byte("forbidden"), wantMemberGone: false},
		{name: "429 does not wrap ErrMemberNotFound", statusCode: http.StatusTooManyRequests, body: []byte("rate limited"), wantMemberGone: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := newStatusError(tt.statusCode, tt.body)
			assert.Error(t, err)
			assert.Equal(t, tt.wantMemberGone, errors.Is(err, ErrMemberNotFound))
			assert.Contains(t, err.Error(), string(tt.body))
		})
	}
}

func TestIsBreakerSuccess(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error is success", err: nil, want: true},
		{name: "404 member-not-found is success", err: newStatusError(http.StatusNotFound, []byte(`{"error":{"code":"not-found","message":""}}`)), want: true},
		{name: "wrapped member-not-found is success", err: fmt.Errorf("request failed: %w", newStatusError(http.StatusNotFound, nil)), want: true},
		{name: "500 is failure", err: newStatusError(http.StatusInternalServerError, []byte("internal error")), want: false},
		{name: "429 is failure", err: newStatusError(http.StatusTooManyRequests, []byte("rate limited")), want: false},
		{name: "generic error is failure", err: errors.New("boom"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsBreakerSuccess(tt.err))
		})
	}
}
