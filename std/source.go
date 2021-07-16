package std

import (
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
)

type (
	SourceRef interface {
		queue.InQueue
		queue.CloserQueue
	}

	sourceImpl struct {
		out       queue.Queue
		graphTree *stream.GraphTree
		left      interface{}
	}
)

var _ stream.Source = (*sourceImpl)(nil)

func NewChannelSource(buffer int) (SourceRef, stream.Source) {
	in := queue.NewQueueEmpty(buffer)
	out := queue.NewQueueEmpty(buffer)
	return in, &sourceImpl{
		out:       out,
		graphTree: stream.PassThrowGraph(in, out),
	}
}

func NewSource(list []interface{}, buffer int) stream.Source {
	in := queue.NewQueueSlice(list)
	out := queue.NewQueueEmpty(buffer)
	return &sourceImpl{
		out:       out,
		graphTree: stream.PassThrowGraph(in, out),
	}
}
func (a *sourceImpl) Via(flow stream.Flow) stream.Source {
	return &sourceImpl{
		out: flow.Out(),
		graphTree: a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	}
}
func (a *sourceImpl) To(sink stream.Sink) stream.Runnable {
	return stream.NewRunnable(
		a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	)
}
func (a *sourceImpl) Out() queue.Queue {
	return a.out
}
func (a *sourceImpl) GraphTree() *stream.GraphTree {
	return a.graphTree
}
