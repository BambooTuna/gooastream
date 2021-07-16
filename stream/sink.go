package stream

import (
	"github.com/BambooTuna/gooastream/queue"
)

type (
	sinkImpl struct {
		in        queue.Queue
		graphTree *GraphTree
	}
)

var _ Sink = (*sinkImpl)(nil)

func BuildSink(in queue.Queue, graphTree *GraphTree) Sink {
	return &sinkImpl{
		in:        in,
		graphTree: graphTree,
	}
}

func NewSink(task func(interface{}) error, buffer int) Sink {
	in := queue.NewQueueEmpty(buffer)
	out := queue.NewQueueSink(task)
	return &sinkImpl{
		in:        in,
		graphTree: PassThrowGraph(in, out),
	}
}
func IgnoreSink() Sink {
	in := queue.NewQueueEmpty(0)
	out := queue.NewQueueSink(func(interface{}) error { return nil })
	return &sinkImpl{
		in:        in,
		graphTree: PassThrowGraph(in, out),
	}
}
func (a *sinkImpl) Dummy() {
}
func (a *sinkImpl) In() queue.Queue {
	return a.in
}
func (a *sinkImpl) GraphTree() *GraphTree {
	return a.graphTree
}
