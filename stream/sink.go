package stream

import (
	"github.com/BambooTuna/gooastream/builder"
	"github.com/BambooTuna/gooastream/queue"
)

type (
	Sink interface {
		Dummy()
		Inlet

		builder.GraphNode
	}

	sinkImpl struct {
		id        string
		in        queue.Queue
		graphTree builder.GraphTree
	}
)

var _ Sink = (*sinkImpl)(nil)

func NewSink(task func(interface{}) error, buffer int) Sink {
	in := queue.NewQueueEmpty(buffer)
	out := queue.NewQueueSink(task)
	return &sinkImpl{
		id:        "sink",
		in:        in,
		graphTree: builder.PassThrowGraph(in, out),
	}
}
func (a *sinkImpl) Dummy() {
}
func (a *sinkImpl) In() queue.Queue {
	return a.in
}
func (a *sinkImpl) GraphTree() builder.GraphTree {
	return a.graphTree
}
