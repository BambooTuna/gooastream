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

type (
	CandidateSourceConfig struct {
		Buffer int
	}
	webrtcCandidateSource struct {
		out       queue.Queue
		graphTree *stream.GraphTree

		conf *CandidateSourceConfig
		peer *webrtc.PeerConnection
	}
)

// webrtc.ICECandidateInit ->
func NewWebrtcCandidateSource(ctx context.Context, conf *CandidateSourceConfig, peer *webrtc.PeerConnection) stream.Source {
	out := queue.NewQueueEmpty(conf.Buffer)
	source := webrtcCandidateSource{
		out:       out,
		graphTree: stream.EmptyGraph(),
		conf:      conf,
		peer:      peer,
	}
	go source.connect(ctx)
	return &source
}

func (a webrtcCandidateSource) Via(flow stream.Flow) stream.Source {
	return stream.BuildSource(
		flow.Out(),
		a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	)
}

func (a webrtcCandidateSource) To(sink stream.Sink) stream.Runnable {
	return stream.NewRunnable(
		a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	)
}

func (a webrtcCandidateSource) Out() queue.Queue {
	return a.out
}

func (a webrtcCandidateSource) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func (a webrtcCandidateSource) connect(ctx context.Context) {
	a.peer.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		_ = a.out.Push(ctx, candidate.ToJSON())
	})
}

var _ stream.Source = (*webrtcCandidateSource)(nil)

type (
	TrackSourceConfig struct {
		ReadDeadline time.Duration
		Buffer       int
	}
	webrtcTrackSource struct {
		out       queue.Queue
		graphTree *stream.GraphTree

		conf *TrackSourceConfig
		peer *webrtc.PeerConnection
	}
)

// *rtp.Packet ->
func NewWebrtcTrackSource(ctx context.Context, conf *TrackSourceConfig, peer *webrtc.PeerConnection) stream.Source {
	out := queue.NewQueueEmpty(conf.Buffer)
	source := webrtcTrackSource{
		out:       out,
		graphTree: stream.EmptyGraph(),
		conf:      conf,
		peer:      peer,
	}
	go source.connect(ctx)
	return &source
}

func (a webrtcTrackSource) Via(flow stream.Flow) stream.Source {
	return stream.BuildSource(
		flow.Out(),
		a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	)
}

func (a webrtcTrackSource) To(sink stream.Sink) stream.Runnable {
	return stream.NewRunnable(
		a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	)
}

func (a webrtcTrackSource) Out() queue.Queue {
	return a.out
}

func (a webrtcTrackSource) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func (a webrtcTrackSource) connect(ctx context.Context) {
	ch := make(chan interface{}, 0)
	defer func() {
		_ = a.peer.Close()
		a.out.Close()
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
			err := a.out.Push(ctx, data)
			if err != nil {
				break T
			}
		}
	}
}

var _ stream.Source = (*webrtcTrackSource)(nil)
