package gortsplib

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/frostyfridge/gortsplib/v4/pkg/base"
)

func TestRTSPTCanonicalAddr(t *testing.T) {
	for _, ca := range []struct {
		name string
		url  string
		addr string
	}{
		{
			"rtspt with port",
			"rtspt://example.com:8554/path",
			"example.com:8554",
		},
		{
			"rtspt without port",
			"rtspt://example.com/path",
			"example.com:554",
		},
		{
			"rtspt ipv6 with port",
			"rtspt://[::1]:8554/path",
			"[::1]:8554",
		},
		{
			"rtspt ipv6 without port",
			"rtspt://[::1]/path",
			"[::1]:554",
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			u, err := base.ParseURL(ca.url)
			require.NoError(t, err)
			
			addr := canonicalAddr(u)
			require.Equal(t, ca.addr, addr)
		})
	}
}
