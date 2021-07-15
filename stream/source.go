package stream

import (
	"github.com/BambooTuna/gooastream/builder"
	"github.com/BambooTuna/gooastream/queue"
)

type (
	SourceRef interface {
		queue.InQueue
		queue.CloserQueue
	}
	Source interface {
		Via(Flow) Source
		To(Sink) Runnable

		Outlet
		builder.GraphNode
	}
	sourceImpl struct {
		out       queue.Queue
		graphTree builder.GraphTree
		left      interface{}
	}
)

var _ Source = (*sourceImpl)(nil)

func NewChannelSource(buffer int) (SourceRef, Source) {
	in := queue.NewQueueEmpty(buffer)
	out := queue.NewQueueEmpty(buffer)
	return in, &sourceImpl{
		out:       out,
		graphTree: builder.PassThrowGraph(in, out),
	}
}

func NewSource(list []interface{}, buffer int) Source {
	in := queue.NewQueueSlice(list)
	out := queue.NewQueueEmpty(buffer)
	return &sourceImpl{
		out:       out,
		graphTree: builder.PassThrowGraph(in, out),
	}
}
func (a *sourceImpl) Via(flow Flow) Source {
	return &sourceImpl{
		out: flow.Out(),
		graphTree: a.GraphTree().
			Append(builder.PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	}
}
func (a *sourceImpl) To(sink Sink) Runnable {
	return &runnableImpl{
		graphTree: a.GraphTree().
			Append(builder.PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	}
}
func (a *sourceImpl) Out() queue.Queue {
	return a.out
}
func (a *sourceImpl) GraphTree() builder.GraphTree {
	return a.graphTree
}
