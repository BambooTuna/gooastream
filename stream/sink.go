package stream

import (
	"github.com/BambooTuna/gooastream/queue"
)

type sinkImpl struct {
	in        queue.Queue
	graphTree *GraphTree
}

var _ Sink = (*sinkImpl)(nil)

/*
	BuildSink
	Create a Sink from input queue.Queue and GraphTree.
	Have one input port and no output port.
*/
func BuildSink(in queue.Queue, graphTree *GraphTree) Sink {
	return &sinkImpl{
		in:        in,
		graphTree: graphTree,
	}
}

/*
	NewSink
	Create a Sink with a function.
	Have one input port and no output port.
	This Sink has no buffer.
*/
func NewSink(task func(interface{}) error) Sink {
	in := queue.NewQueueEmpty(0)
	out := queue.NewQueueSink(task)
	return &sinkImpl{
		in:        in,
		graphTree: PassThrowGraph(in, out),
	}
}

/*
	IgnoreSink
	Create a Sink.
	Have one input port and no output port.
	This Sink just throws away the data it receives.
*/
func IgnoreSink() Sink {
	in := queue.NewQueueEmpty(0)
	out := queue.NewQueueSink(func(interface{}) error { return nil })
	return &sinkImpl{
		in:        in,
		graphTree: PassThrowGraph(in, out),
	}
}

/*
	Dummy
	It's just to fill the interface, it doesn't make sense.
*/
func (a *sinkImpl) Dummy() {}

/*
	In
	queue.Queue for upstream input.
	Should not be used.
*/
func (a *sinkImpl) In() queue.Queue {
	return a.in
}

/*
	GraphTree
	Represents the connection between queue.Queue.
	Should not be used.
*/
func (a *sinkImpl) GraphTree() *GraphTree {
	return a.graphTree
}
