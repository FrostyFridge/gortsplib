//go:build cgo

package main

import (
	"crypto/rand"
	"log"
	"time"

	"github.com/frostyfridge/gortsplib/v4"
	"github.com/frostyfridge/gortsplib/v4/pkg/description"
	"github.com/frostyfridge/gortsplib/v4/pkg/format"
)

// This example shows how to
// 1. connect to a RTSP server, announce an AV1 format.
// 2. generate dummy RGBA images.
// 3. encode images with AV1.
// 4. generate RTP packets from AV1.
// 5. write RTP packets to the server.

// This example requires the FFmpeg libraries, that can be installed with this command:
// apt install -y libavcodec-dev libswscale-dev gcc pkg-config

func multiplyAndDivide(v, m, d int64) int64 {
	secs := v / d
	dec := v % d
	return (secs*m + dec*m/d)
}

func randUint32() (uint32, error) {
	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]), nil
}

func main() {
	// create a stream description that contains a AV1 format
	forma := &format.AV1{
		PayloadTyp: 96,
	}
	desc := &description.Session{
		Medias: []*description.Media{{
			Type:    description.MediaTypeVideo,
			Formats: []format.Format{forma},
		}},
	}

	// connect to the server, announce the format and start recording
	c := gortsplib.Client{}
	err := c.StartRecording("rtsp://myuser:mypass@localhost:8554/mystream", desc)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// setup RGBA -> AV1 encoder
	av1enc := &av1Encoder{
		Width:  640,
		Height: 480,
		FPS:    5,
	}
	err = av1enc.initialize()
	if err != nil {
		panic(err)
	}
	defer av1enc.close()

	// setup AV1 -> RTP encoder
	rtpEnc, err := forma.CreateEncoder()
	if err != nil {
		panic(err)
	}

	start := time.Now()

	randomStart, err := randUint32()
	if err != nil {
		panic(err)
	}

	// setup a ticker to sleep between frames
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		// get current timestamp
		pts := multiplyAndDivide(int64(time.Since(start)), int64(forma.ClockRate()), int64(time.Second))

		// create a dummy image
		img := createDummyImage()

		// encode the image with AV1
		au, pts, err := av1enc.encode(img, pts)
		if err != nil {
			panic(err)
		}

		// wait for a AV1 access unit
		if au == nil {
			continue
		}

		// generate RTP packets from the AV1 access unit
		pkts, err := rtpEnc.Encode(au)
		if err != nil {
			panic(err)
		}

		log.Printf("writing RTP packets with PTS=%d, au=%d, pkts=%d", pts, len(au), len(pkts))

		// write RTP packets to the server
		for _, pkt := range pkts {
			pkt.Timestamp += uint32(int64(randomStart) + pts)

			err = c.WritePacketRTP(desc.Medias[0], pkt)
			if err != nil {
				panic(err)
			}
		}
	}
}
