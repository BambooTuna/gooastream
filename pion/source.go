package pion

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"strings"
	"time"
)

type CandidateSourceConfig struct {
	Buffer int
}

// webrtc.ICECandidateInit ->
func NewWebrtcCandidateSource(conf *CandidateSourceConfig, peer *webrtc.PeerConnection) stream.Source {
	out := queue.NewQueueEmpty(conf.Buffer)
	graphTree := stream.EmptyGraph()
	graphTree.AddWire(newWebrtcCandidateSourceWire(out, conf, peer))
	return stream.BuildSource(out, graphTree)
}

type webrtcCandidateSourceWire struct {
	to queue.InQueue

	conf *CandidateSourceConfig
	peer *webrtc.PeerConnection
}

func newWebrtcCandidateSourceWire(to queue.Queue, conf *CandidateSourceConfig, peer *webrtc.PeerConnection) stream.Wire {
	return &webrtcCandidateSourceWire{
		to:   to,
		conf: conf,
		peer: peer,
	}
}

func (a webrtcCandidateSourceWire) Run(ctx context.Context, cancel context.CancelFunc) {
	a.peer.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		_ = a.to.Push(ctx, candidate.ToJSON())
	})
}

var _ stream.Wire = (*webrtcCandidateSourceWire)(nil)

type TrackSourceConfig struct {
	ReadDeadline time.Duration
	Buffer       int
}

// *rtp.Packet ->
func NewWebrtcTrackSource(conf *TrackSourceConfig, peer *webrtc.PeerConnection) stream.Source {
	out := queue.NewQueueEmpty(conf.Buffer)
	graphTree := stream.EmptyGraph()
	graphTree.AddWire(newWebrtcTrackSourceWire(out, conf, peer))
	return stream.BuildSource(out, graphTree)
}

type webrtcTrackSourceWire struct {
	to queue.InQueue

	conf *TrackSourceConfig
	peer *webrtc.PeerConnection
}

func newWebrtcTrackSourceWire(to queue.InQueue, conf *TrackSourceConfig, peer *webrtc.PeerConnection) stream.Wire {
	return &webrtcTrackSourceWire{
		to:   to,
		conf: conf,
		peer: peer,
	}
}

func (a webrtcTrackSourceWire) Run(ctx context.Context, cancel context.CancelFunc) {
	ch := make(chan interface{}, 0)
	defer func() {
		cancel()
		_ = a.peer.Close()
		a.to.Close()
	}()
	a.peer.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := a.peer.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())}})
				if errSend != nil {
					break
				}
			}
			ticker.Stop()
			_ = a.peer.Close()
		}()
		codec := remoteTrack.Codec()
		if strings.EqualFold(codec.MimeType, webrtc.MimeTypeOpus) {
			for {
				if err := remoteTrack.SetReadDeadline(time.Now().Add(a.conf.ReadDeadline)); err != nil {
					break
				}
				packet, _, err := remoteTrack.ReadRTP()
				if err != nil {
					break
				}
				ch <- packet
			}
		}
		close(ch)
	})
T:
	for {
		select {
		case <-ctx.Done():
			break T
		case data, ok := <-ch:
			if !ok {
				break T
			}
			err := a.to.Push(ctx, data)
			if err != nil {
				break T
			}
		}
	}
}

var _ stream.Wire = (*webrtcTrackSourceWire)(nil)
