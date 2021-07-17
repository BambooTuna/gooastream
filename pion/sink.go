package pion

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type CandidateSinkConfig struct {
	Buffer int
}

type webrtcCandidateSink struct {
	in        queue.Queue
	graphTree *stream.GraphTree

	conf *CandidateSinkConfig
	peer *webrtc.PeerConnection
}

var _ stream.Sink = (*webrtcCandidateSink)(nil)

// -> webrtc.ICECandidateInit
func NewWebrtcCandidateSink(ctx context.Context, conf *CandidateSinkConfig, peer *webrtc.PeerConnection) stream.Sink {
	in := queue.NewQueueEmpty(conf.Buffer)
	sink := webrtcCandidateSink{
		in:        in,
		graphTree: stream.EmptyGraph(),

		conf: conf,
		peer: peer,
	}
	go sink.connect(ctx)
	return &sink
}

func (a webrtcCandidateSink) Dummy() {
}

func (a webrtcCandidateSink) In() queue.Queue {
	return a.in
}

func (a webrtcCandidateSink) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func (a webrtcCandidateSink) connect(ctx context.Context) {
	defer func() {
		a.in.Close()
		_ = a.peer.Close()
	}()
	for {
		data, err := a.in.Pop(ctx)
		if err != nil {
			break
		}
		candidate, ok := data.(webrtc.ICECandidateInit)
		if !ok {
			continue
		}
		_ = a.peer.AddICECandidate(candidate)
	}
}

type TrackSinkConfig struct {
	Transceiver  *webrtc.RTPTransceiver
	DefaultTrack *webrtc.TrackLocalStaticRTP
	Buffer       int
}

type webrtcTrackSink struct {
	in        queue.Queue
	graphTree *stream.GraphTree

	conf  *TrackSinkConfig
	peer  *webrtc.PeerConnection
	track *webrtc.TrackLocalStaticRTP
}

var _ stream.Sink = (*webrtcTrackSink)(nil)

// -> *rtp.Packet
func NewWebrtcTrackSink(ctx context.Context, conf *TrackSinkConfig, peer *webrtc.PeerConnection) (stream.Sink, error) {
	in := queue.NewQueueEmpty(conf.Buffer)
	track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2, SDPFmtpLine: "", RTCPFeedback: nil}, conf.DefaultTrack.ID(), conf.DefaultTrack.StreamID())
	if err != nil {
		return nil, err
	}
	err = conf.Transceiver.Sender().ReplaceTrack(track)
	if err != nil {
		return nil, err
	}
	sink := webrtcTrackSink{
		in:        in,
		graphTree: stream.EmptyGraph(),

		conf:  conf,
		peer:  peer,
		track: track,
	}
	go sink.connect(ctx)
	return &sink, nil
}

func (a webrtcTrackSink) Dummy() {
}

func (a webrtcTrackSink) In() queue.Queue {
	return a.in
}

func (a webrtcTrackSink) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func (a webrtcTrackSink) connect(ctx context.Context) {
	defer func() {
		a.in.Close()
		_ = a.conf.Transceiver.Sender().ReplaceTrack(a.conf.DefaultTrack)
	}()
	for {
		data, err := a.in.Pop(ctx)
		if err != nil {
			break
		}
		packet, ok := data.(*rtp.Packet)
		if !ok {
			continue
		}
		err = a.track.WriteRTP(packet)
		if err != nil {
			break
		}
	}
}
