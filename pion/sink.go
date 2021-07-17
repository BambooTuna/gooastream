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

// -> webrtc.ICECandidateInit
func NewWebrtcCandidateSink(conf *CandidateSinkConfig, peer *webrtc.PeerConnection) stream.Sink {
	in := queue.NewQueueEmpty(conf.Buffer)
	graphTree := stream.EmptyGraph()
	graphTree.AddWire(newWebrtcCandidateSinkWire(in, conf, peer))
	return stream.BuildSink(in, graphTree)
}

type webrtcCandidateSinkWire struct {
	from queue.OutQueue

	conf *CandidateSinkConfig
	peer *webrtc.PeerConnection
}

func newWebrtcCandidateSinkWire(from queue.OutQueue, conf *CandidateSinkConfig, peer *webrtc.PeerConnection) stream.Wire {
	return &webrtcCandidateSinkWire{
		from: from,
		conf: conf,
		peer: peer,
	}
}

func (a webrtcCandidateSinkWire) Run(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		cancel()
		a.from.Close()
		_ = a.peer.Close()
	}()
T:
	for {
		select {
		case <-ctx.Done():
			break T
		default:
			data, err := a.from.Pop(ctx)
			if err != nil {
				break T
			}
			candidate, ok := data.(webrtc.ICECandidateInit)
			if !ok {
				continue
			}
			_ = a.peer.AddICECandidate(candidate)
		}
	}
}

var _ stream.Wire = (*webrtcCandidateSinkWire)(nil)

type TrackSinkConfig struct {
	Transceiver  *webrtc.RTPTransceiver
	DefaultTrack *webrtc.TrackLocalStaticRTP
	Buffer       int
}

// -> *rtp.Packet
// -> []byte
func NewWebrtcTrackSink(conf *TrackSinkConfig, peer *webrtc.PeerConnection) (stream.Sink, error) {
	track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2, SDPFmtpLine: "", RTCPFeedback: nil}, conf.DefaultTrack.ID(), conf.DefaultTrack.StreamID())
	if err != nil {
		return nil, err
	}
	err = conf.Transceiver.Sender().ReplaceTrack(track)
	if err != nil {
		return nil, err
	}
	in := queue.NewQueueEmpty(conf.Buffer)
	graphTree := stream.EmptyGraph()
	graphTree.AddWire(newWebrtcTrackSinkWire(in, conf, peer, track))
	return stream.BuildSink(in, graphTree), nil
}

type webrtcTrackSinkWire struct {
	to queue.Queue

	conf  *TrackSinkConfig
	peer  *webrtc.PeerConnection
	track *webrtc.TrackLocalStaticRTP
}

func newWebrtcTrackSinkWire(to queue.Queue, conf *TrackSinkConfig, peer *webrtc.PeerConnection, track *webrtc.TrackLocalStaticRTP) stream.Wire {
	return &webrtcTrackSinkWire{
		to:    to,
		conf:  conf,
		peer:  peer,
		track: track,
	}
}

func (a webrtcTrackSinkWire) Run(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		cancel()
		a.to.Close()
		_ = a.conf.Transceiver.Sender().ReplaceTrack(a.conf.DefaultTrack)
	}()
T:
	for {
		select {
		case <-ctx.Done():
			break T
		default:
			data, err := a.to.Pop(ctx)
			if err != nil {
				break T
			}
			switch packet := data.(type) {
			case *rtp.Packet:
				err = a.track.WriteRTP(packet)
				if err != nil {
					break T
				}
			case []byte:
				_, err = a.track.Write(packet)
				if err != nil {
					break T
				}
			default:
				continue
			}
		}
	}
}

var _ stream.Wire = (*webrtcTrackSinkWire)(nil)
