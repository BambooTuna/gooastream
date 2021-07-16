package std

import (
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
)

type (
	sinkImpl struct {
		in        queue.Queue
		graphTree *stream.GraphTree
	}
)

var _ stream.Sink = (*sinkImpl)(nil)

func NewSink(task func(interface{}) error, buffer int) stream.Sink {
	in := queue.NewQueueEmpty(buffer)
	out := queue.NewQueueSink(task)
	return &sinkImpl{
		in:        in,
		graphTree: stream.PassThrowGraph(in, out),
	}
}
func IgnoreSink() stream.Sink {
	in := queue.NewQueueEmpty(0)
	out := queue.NewQueueSink(func(interface{}) error { return nil })
	return &sinkImpl{
		in:        in,
		graphTree: stream.PassThrowGraph(in, out),
	}
}
func (a *sinkImpl) Dummy() {
}
func (a *sinkImpl) In() queue.Queue {
	return a.in
}
func (a *sinkImpl) GraphTree() *stream.GraphTree {
	return a.graphTree
}
