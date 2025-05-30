package main

import (
	"crypto/tls"
	"log"
	"sync"

	"github.com/pion/rtp"

	"github.com/frostyfridge/gortsplib/v4"
	"github.com/frostyfridge/gortsplib/v4/pkg/base"
	"github.com/frostyfridge/gortsplib/v4/pkg/description"
	"github.com/frostyfridge/gortsplib/v4/pkg/format"
)

// This example shows how to
// 1. create a RTSP server which accepts only connections encrypted with TLS (RTSPS).
// 2. allow a single client to publish a stream with TCP.
// 3. allow multiple clients to read the stream with TCP.

type serverHandler struct {
	server    *gortsplib.Server
	mutex     sync.RWMutex
	stream    *gortsplib.ServerStream
	publisher *gortsplib.ServerSession
}

// called when a connection is opened.
func (sh *serverHandler) OnConnOpen(ctx *gortsplib.ServerHandlerOnConnOpenCtx) {
	log.Printf("conn opened")
}

// called when a connection is closed.
func (sh *serverHandler) OnConnClose(ctx *gortsplib.ServerHandlerOnConnCloseCtx) {
	log.Printf("conn closed (%v)", ctx.Error)
}

// called when a session is opened.
func (sh *serverHandler) OnSessionOpen(ctx *gortsplib.ServerHandlerOnSessionOpenCtx) {
	log.Printf("session opened")
}

// called when a session is closed.
func (sh *serverHandler) OnSessionClose(ctx *gortsplib.ServerHandlerOnSessionCloseCtx) {
	log.Printf("session closed")

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	// if the session is the publisher,
	// close the stream and disconnect any reader.
	if sh.stream != nil && ctx.Session == sh.publisher {
		sh.stream.Close()
		sh.stream = nil
	}
}

// called when receiving a DESCRIBE request.
func (sh *serverHandler) OnDescribe(ctx *gortsplib.ServerHandlerOnDescribeCtx) (*base.Response, *gortsplib.ServerStream, error) {
	log.Printf("DESCRIBE request")

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	// no one is publishing yet
	if sh.stream == nil {
		return &base.Response{
			StatusCode: base.StatusNotFound,
		}, nil, nil
	}

	// send medias that are being published to the client
	return &base.Response{
		StatusCode: base.StatusOK,
	}, sh.stream, nil
}

// called when receiving an ANNOUNCE request.
func (sh *serverHandler) OnAnnounce(ctx *gortsplib.ServerHandlerOnAnnounceCtx) (*base.Response, error) {
	log.Printf("ANNOUNCE request")

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	// disconnect existing publisher
	if sh.stream != nil {
		sh.stream.Close()
		sh.publisher.Close()
	}

	// create the stream and save the publisher
	sh.stream = &gortsplib.ServerStream{
		Server: sh.server,
		Desc:   ctx.Description,
	}
	err := sh.stream.Initialize()
	if err != nil {
		panic(err)
	}
	sh.publisher = ctx.Session

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// called when receiving a SETUP request.
func (sh *serverHandler) OnSetup(ctx *gortsplib.ServerHandlerOnSetupCtx) (*base.Response, *gortsplib.ServerStream, error) {
	log.Printf("SETUP request")

	// SETUP is used by both readers and publishers. In case of publishers, just return StatusOK.
	if ctx.Session.State() == gortsplib.ServerSessionStatePreRecord {
		return &base.Response{
			StatusCode: base.StatusOK,
		}, nil, nil
	}

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	// no one is publishing yet
	if sh.stream == nil {
		return &base.Response{
			StatusCode: base.StatusNotFound,
		}, nil, nil
	}

	return &base.Response{
		StatusCode: base.StatusOK,
	}, sh.stream, nil
}

// called when receiving a PLAY request.
func (sh *serverHandler) OnPlay(ctx *gortsplib.ServerHandlerOnPlayCtx) (*base.Response, error) {
	log.Printf("PLAY request")

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// called when receiving a RECORD request.
func (sh *serverHandler) OnRecord(ctx *gortsplib.ServerHandlerOnRecordCtx) (*base.Response, error) {
	log.Printf("RECORD request")

	// called when receiving a RTP packet
	ctx.Session.OnPacketRTPAny(func(medi *description.Media, forma format.Format, pkt *rtp.Packet) {
		// route the RTP packet to all readers
		sh.stream.WritePacketRTP(medi, pkt)
	})

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

func main() {
	// setup certificates - they can be generated with
	// openssl genrsa -out server.key 2048
	// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		panic(err)
	}

	// configure the server
	h := &serverHandler{}
	h.server = &gortsplib.Server{
		Handler:     h,
		TLSConfig:   &tls.Config{Certificates: []tls.Certificate{cert}},
		RTSPAddress: ":8322",
	}

	// start server and wait until a fatal error
	log.Printf("server is ready on %s", h.server.RTSPAddress)
	panic(h.server.StartAndWait())
}
