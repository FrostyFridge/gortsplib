package gortsplib

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/frostyfridge/gortsplib/v4/pkg/base"
	"github.com/frostyfridge/gortsplib/v4/pkg/liberrors"
)

func TestClientTransportSecurityForced(t *testing.T) {
	// Test cases to ensure certain schemes enforce TCP transport
	for _, tc := range []struct {
		name     string
		scheme   string
		isSecure bool
	}{
		{
			name:     "RTSPS enforces TCP",
			scheme:   "rtsps",
			isSecure: true,
		},
		{
			name:     "RTSPT enforces TCP",
			scheme:   "rtspt",
			isSecure: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &Client{}

			// Try using a non-TCP transport with secure schemes
			udpTransport := TransportUDP
			client.Transport = &udpTransport

			err := client.Start(tc.scheme, "example.com:554")
			require.NoError(t, err) // Start itself doesn't fail

			// Mock the connection URL since we don't actually connect
			client.connURL = &base.URL{
				Scheme: tc.scheme,
				Host:   "example.com:554",
			}

			// connOpen should fail because non-TCP transport isn't allowed with secure schemes
			err = client.connOpen()
			require.Error(t, err)
			errMsg := err.Error()
			require.Contains(t, errMsg, "can't be used with a non-TCP transport protocol")
		})
	}

	// Regular RTSP can use UDP
	t.Run("RTSP allows UDP", func(t *testing.T) {
		client := &Client{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				// Mock successful connection
				return nil, fmt.Errorf("expected error")
			},
		}

		// Setup UDP transport
		udpTransport := TransportUDP
		client.Transport = &udpTransport

		err := client.Start("rtsp", "example.com:554")
		require.NoError(t, err)

		// Mock the connection URL
		client.connURL = &base.URL{
			Scheme: "rtsp",
			Host:   "example.com:554",
		}

		// This should error with connection error, not transport error
		err = client.connOpen()
		require.Error(t, err)
		_, ok := err.(liberrors.ErrClientRTSPSTCP)
		require.False(t, ok, "Error should not be ErrClientRTSPSTCP")
	})
}
