package gortsplib

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/frostyfridge/gortsplib/v4/pkg/base"
)

func TestClientStartWithRTSPT(t *testing.T) {
	// Test the full client Start method with rtspt
	cases := []struct {
		name   string
		scheme string
		wantTCP bool
	}{
		{"RTSP doesn't force TCP", "rtsp", false},
		{"RTSPS forces TCP", "rtsps", true},
		{"RTSPT forces TCP", "rtspt", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Client{}
			
			// Start client with the scheme
			err := c.Start(tc.scheme, "example.com:554")
			require.NoError(t, err)
			
			if tc.wantTCP {
				// If TCP should be forced
				require.NotNil(t, c.Transport, "Transport should be set")
				require.Equal(t, TransportTCP, *c.Transport, 
					"Transport should be TCP for scheme: "+tc.scheme)
			} else {
				// Regular RTSP shouldn't set Transport
				if c.Transport != nil {
					require.NotEqual(t, TransportTCP, *c.Transport, 
						"Transport should not be TCP for scheme: "+tc.scheme)
				}
			}
			
			// Test URL parsing for the scheme
			u, err := base.ParseURL(tc.scheme+"://example.com:554/stream")
			require.NoError(t, err)
			require.Equal(t, tc.scheme, u.Scheme)
			
			// Test canonical address
			addr := canonicalAddr(u)
			require.Equal(t, "example.com:554", addr)
		})
	}
}
