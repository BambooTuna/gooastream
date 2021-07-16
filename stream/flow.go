package stream

import (
	"github.com/BambooTuna/gooastream/queue"
)

type flowImpl struct {
	in        queue.Queue
	out       queue.Queue
	graphTree *GraphTree
}

var _ Flow = (*flowImpl)(nil)

/*
	BuildFlow
	Create a Flow from input and output queue.Queue and GraphTree.
	Have one input port and one output port.
*/
func BuildFlow(in, out queue.Queue, graphTree *GraphTree) Flow {
	return &flowImpl{
		in:        in,
		out:       out,
		graphTree: graphTree,
	}
}

/*
	NewBufferFlow
	Create a Flow with a buffer.
	Have one input port and one output port.
	If the downstream is clogged, it will accumulate in the buffer.
*/
func NewBufferFlow(buffer int) Flow {
	in := queue.NewQueueEmpty(buffer)
	out := queue.NewQueueEmpty(buffer)
	return &flowImpl{
		in:        in,
		out:       out,
		graphTree: PassThrowGraph(in, out),
	}
}

/*
	FlowFromSinkAndSource
	Create a Flow from Sink and Source.
	Have one input port and one output port.
	Even if the downstream is clogged, it does not affect the upstream.
     +-------------------------------------------------------+
     | Resulting Flow                                       |
     |                                                      |
     |  +-------------+                  +---------------+  |
     |  |             |                  |               |  |
 I  ~~> | Sink        | [no-connection!] | Source        | ~~> O
     |  |             |                  |               |  |
     |  +-------------+                  +---------------+  |
     +------------------------------------------------------+
*/
func FlowFromSinkAndSource(sink Sink, source Source) Flow {
	return &flowImpl{
		in:  sink.In(),
		out: source.Out(),
		graphTree: sink.GraphTree().
			Append(source.GraphTree()),
	}
}

/*
	Via
	Create one Flow by connecting another Flow to the downstream.
	Have one input port and one output port.
     +-------------------------------------------------------+
     | Resulting Flow                                       |
     |                                                      |
     |  +-------------+                  +---------------+  |
     |  |             |                  |               |  |
 I  ~~> | Flow        |       ~~~>       | Flow          | ~~> O
     |  |             |                  |               |  |
     |  +-------------+                  +---------------+  |
     +------------------------------------------------------+
*/
func (a *flowImpl) Via(flow Flow) Flow {
	return &flowImpl{
		in:  a.In(),
		out: flow.Out(),
		graphTree: a.GraphTree().
			Append(PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	}
}

/*
	To
	Create one Sink by connecting another Sink to the downstream.
	Have one input port and no output port.
     +-------------------------------------------------------+
     | Resulting Sink                                       |
     |                                                      |
     |  +-------------+                  +---------------+  |
     |  |             |                  |               |  |
 I  ~~> | Flow        |       ~~~>       | Sink          |  |
     |  |             |                  |               |  |
     |  +-------------+                  +---------------+  |
     +------------------------------------------------------+
*/
func (a *flowImpl) To(sink Sink) Sink {
	return &sinkImpl{
		in: a.In(),
		graphTree: a.GraphTree().
			Append(PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	}
}

/*
	In
	queue.Queue for upstream input.
	Should not be used.
*/
func (a *flowImpl) In() queue.Queue {
	return a.in
}

/*
	Out
	queue.Queue for downstream output.
	Should not be used.
*/
func (a *flowImpl) Out() queue.Queue {
	return a.out
}

/*
	GraphTree
	Represents the connection between queue.Queue.
	Should not be used.
*/
func (a *flowImpl) GraphTree() *GraphTree {
	return a.graphTree
}
