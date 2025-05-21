package gortsplib

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/frostyfridge/gortsplib/v4/pkg/base"
	"github.com/frostyfridge/gortsplib/v4/pkg/conn"
)

func TestClientRtsptEnforcesTCP(t *testing.T) {
	// Create a listener for testing
	l, err := net.Listen("tcp", "localhost:8554")
	require.NoError(t, err)
	defer l.Close()

	serverDone := make(chan struct{})
	defer func() { <-serverDone }()

	// Start a goroutine that pretends to be a server
	go func() {
		defer close(serverDone)

		nconn, err := l.Accept()
		require.NoError(t, err)
		defer nconn.Close()

		// Create a conn object
		connObj := conn.NewConn(nconn)

		// Read the OPTIONS request
		req, err := connObj.ReadRequest()
		require.NoError(t, err)

		// Verify it came from an rtspt URL
		require.Equal(t, base.Options, req.Method)
		require.Equal(t, "rtspt://localhost:8554/", req.URL.String())

		// Send a response
		err = connObj.WriteResponse(&base.Response{
			StatusCode: base.StatusOK,
			Header: base.Header{
				"CSeq": base.HeaderValue{"1"},
			},
		})
		require.NoError(t, err)
	}()

	// Create a client
	client := &Client{}

	// Set a short timeout for test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dialDone := make(chan struct{})

	// Initialize the client with a special dial function to capture the URL
	client.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		defer close(dialDone)
		return net.Dial(network, address)
	}

	// Start the client with the rtspt scheme
	err = client.Start("rtspt", "localhost:8554")
	require.NoError(t, err)
	defer client.Close()

	// Wait for client to establish connection
	select {
	case <-dialDone:
	case <-ctx.Done():
		t.Fatal("timeout waiting for client to connect")
	}

	// Make an OPTIONS request to test the connection
	_, err = client.Options(&base.URL{
		Scheme: "rtspt",
		Host:   "localhost:8554",
		Path:   "/",
	})
	require.NoError(t, err)

	// Verify that TCP transport is forced
	require.NotNil(t, client.Transport)
	require.Equal(t, TransportTCP, *client.Transport)
}

func TestClientRtsptRequiresTCP(t *testing.T) {
	// This test will be covered by the more comprehensive TestClientTransportSecurityForced
	t.Skip("This test is covered by TestClientTransportSecurityForced")
}
