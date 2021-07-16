package stream

import (
	"github.com/BambooTuna/gooastream/queue"
)

type (
	SourceRef interface {
		queue.InQueue
		queue.CloserQueue
	}

	sourceImpl struct {
		out       queue.Queue
		graphTree *GraphTree
		left      interface{}
	}
)

var _ Source = (*sourceImpl)(nil)

func BuildSource(out queue.Queue, graphTree *GraphTree) Source {
	return &sourceImpl{
		out:       out,
		graphTree: graphTree,
	}
}

func NewChannelSource(buffer int) (SourceRef, Source) {
	in := queue.NewQueueEmpty(buffer)
	out := queue.NewQueueEmpty(buffer)
	return in, &sourceImpl{
		out:       out,
		graphTree: PassThrowGraph(in, out),
	}
}

func NewSource(list []interface{}, buffer int) Source {
	in := queue.NewQueueSlice(list)
	out := queue.NewQueueEmpty(buffer)
	return &sourceImpl{
		out:       out,
		graphTree: PassThrowGraph(in, out),
	}
}
func (a *sourceImpl) Via(flow Flow) Source {
	return &sourceImpl{
		out: flow.Out(),
		graphTree: a.GraphTree().
			Append(PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	}
}
func (a *sourceImpl) To(sink Sink) Runnable {
	return NewRunnable(
		a.GraphTree().
			Append(PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	)
}
func (a *sourceImpl) Out() queue.Queue {
	return a.out
}
func (a *sourceImpl) GraphTree() *GraphTree {
	return a.graphTree
}
