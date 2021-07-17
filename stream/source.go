package stream

import (
	"github.com/BambooTuna/gooastream/queue"
)

/*
	SourceChannel
	This is input channel.
	If this is closed, it will be transmitted to the entire stream.
*/
type SourceChannel interface {
	queue.InQueue
	queue.CloserQueue
}

type sourceImpl struct {
	out       queue.Queue
	graphTree *GraphTree
}

var _ Source = (*sourceImpl)(nil)

/*
	BuildSource
	Create a Source from output queue.Queue and GraphTree.
	Have no input port and one output port.
*/
func BuildSource(out queue.Queue, graphTree *GraphTree) Source {
	return &sourceImpl{
		out:       out,
		graphTree: graphTree,
	}
}

/*
	NewChannelSource
	Create SourceChannel and Source with a buffer.
	Have no input port and one output port.
	If the downstream is clogged, what you send to SourceChannel will be stored in the buffer.
*/
func NewChannelSource(options ...queue.Option) (SourceChannel, Source) {
	out := queue.NewQueueEmpty(options...)
	return out, &sourceImpl{
		out:       out,
		graphTree: EmptyGraph(),
	}
}

/*
	NewInfiniteSource
	Create a Source with a element.
	Have no input port and one output port.
	Infinitely pass elements downstream.
*/
func NewInfiniteSource(element interface{}) Source {
	out := queue.NewInfiniteElement(element)
	return &sourceImpl{
		out:       out,
		graphTree: EmptyGraph(),
	}
}

/*
	NewSliceSource
	Create a Source with a buffer.
	Have no input port and one output port.
*/
func NewSliceSource(slice []interface{}, options ...queue.Option) Source {
	in := queue.NewQueueSlice(slice)
	out := queue.NewQueueEmpty(options...)
	return &sourceImpl{
		out:       out,
		graphTree: PassThrowGraph(in, out),
	}
}

/*
	Via
	Create one Source by connecting another Flow to the downstream.
	Have no input port and one output port.
     +-------------------------------------------------------+
     | Resulting Source                                     |
     |                                                      |
     |  +-------------+                  +---------------+  |
     |  |             |                  |               |  |
     |  | Source      |       ~~~>       | Flow          | ~~> O
     |  |             |                  |               |  |
     |  +-------------+                  +---------------+  |
     +------------------------------------------------------+
*/
func (a *sourceImpl) Via(flow Flow) Source {
	return &sourceImpl{
		out: flow.Out(),
		graphTree: a.GraphTree().
			Append(PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	}
}

/*
	To
	Create one Runnable by connecting another Sink to the downstream.
	Have no input port and no output port.
     +-------------------------------------------------------+
     | Resulting Runnable                                   |
     |                                                      |
     |  +-------------+                  +---------------+  |
     |  |             |                  |               |  |
     |  | Source      |       ~~~>       | Sink          |  |
     |  |             |                  |               |  |
     |  +-------------+                  +---------------+  |
     +------------------------------------------------------+
*/
func (a *sourceImpl) To(sink Sink) Runnable {
	return NewRunnable(
		a.GraphTree().
			Append(PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	)
}

/*
	Out
	queue.Queue for downstream output.
	Should not be used.
*/
func (a *sourceImpl) Out() queue.Queue {
	return a.out
}

/*
	GraphTree
	Represents the connection between queue.Queue.
	Should not be used.
*/
func (a *sourceImpl) GraphTree() *GraphTree {
	return a.graphTree
}
