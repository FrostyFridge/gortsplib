package gortsplib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRTSPTForcesTransport(t *testing.T) {
	// Test that when a client is created with an rtspt URL, it forces TCP transport
	client := &Client{}
	
	// Start with rtspt scheme
	err := client.Start("rtspt", "example.com:554")
	require.NoError(t, err)
	
	// Verify Transport is set to TCP
	require.NotNil(t, client.Transport, "Transport should be set")
	require.Equal(t, TransportTCP, *client.Transport, "Transport should be TCP")
	
	// Create another client with rtsp scheme
	client2 := &Client{}
	
	// Start with regular rtsp scheme
	err = client2.Start("rtsp", "example.com:554")
	require.NoError(t, err)
	
	// Verify Transport is not set for regular rtsp (default behavior)
	require.Nil(t, client2.Transport, "Transport should not be set for regular rtsp")
	
	// Create a client with explicit udp transport
	client3 := &Client{}
	udpTransport := TransportUDP
	client3.Transport = &udpTransport
	
	// Start with rtspt scheme
	err = client3.Start("rtspt", "example.com:554")
	require.NoError(t, err)
	
	// Verify Transport was overridden to TCP
	require.NotNil(t, client3.Transport, "Transport should be set")
	require.Equal(t, TransportTCP, *client3.Transport, "Transport should be TCP even if UDP was requested")
}
