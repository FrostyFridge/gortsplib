// Package main contains an example of a RTSP client that connects to a RTSP server,
// and forces TCP transport by using the rtspt:// scheme.
package main

import (
	"log"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"

	"github.com/frostyfridge/gortsplib/v4"
	"github.com/frostyfridge/gortsplib/v4/pkg/base"
	"github.com/frostyfridge/gortsplib/v4/pkg/description"
	"github.com/frostyfridge/gortsplib/v4/pkg/format"
)

// This example shows how to:
// 1. connect to a RTSP server that requires TCP transport
// 2. get and print stream tracks
// 3. receive RTP packets with TCP transport

func main() {
	// Create a client
	c := &gortsplib.Client{}

	// Parse the URL and setup a connection with the server using TCP
	u, err := base.ParseURL("rtspt://example.com:8554/mystream")
	if err != nil {
		panic(err)
	}

	err = c.Start(u.Scheme, u.Host)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// Get a session description
	desc, _, err := c.Describe(u)
	if err != nil {
		panic(err)
	}

	// Setup all medias
	err = c.SetupAll(u, desc.Medias)
	if err != nil {
		panic(err)
	}

	// Called when a RTP packet arrives
	c.OnPacketRTPAny(func(medi *description.Media, forma format.Format, pkt *rtp.Packet) {
		log.Printf("RTP packet from media %v, payload type %v\n", medi, forma)
	})

	// Called when a RTCP packet arrives
	c.OnPacketRTCPAny(func(medi *description.Media, pkt rtcp.Packet) {
		log.Printf("RTCP packet from media %v, %T\n", medi, pkt)
	})

	// Start playing
	_, err = c.Play(nil)
	if err != nil {
		panic(err)
	}

	// Wait until a fatal error
	panic(c.Wait())
}
