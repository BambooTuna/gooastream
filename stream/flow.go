package stream

import (
	"github.com/BambooTuna/gooastream/queue"
	"time"
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
func NewBufferFlow(options ...queue.Option) Flow {
	return NewThrottleFlow(0, options...)
}

/*
	NewThrottleFlow
	Create a Flow with a buffer.
	Have one input port and one output port.
	Once in a certain period pass the data between upstream and downstream.
	If throttle time.Duration is 0 or less, the behavior is the same as pass-through.
*/
func NewThrottleFlow(throttle time.Duration, options ...queue.Option) Flow {
	in := queue.NewQueueEmpty(options...)
	out := queue.NewQueueEmpty(options...)
	return &flowImpl{
		in:        in,
		out:       out,
		graphTree: ThrottleGraph(in, out, throttle),
	}
}

/*
	NewMapFlow
	Create a Flow with a buffer and map function.
	Have one input port and one output port.
	Pass the result of passing the data of upstream through the function to downstream.
*/
func NewMapFlow(f func(interface{}) (interface{}, error), options ...queue.Option) Flow {
	in := queue.NewQueueEmpty(options...)
	out := queue.NewQueueEmpty(options...)
	return &flowImpl{
		in:        in,
		out:       out,
		graphTree: MapGraph(in, out, f),
	}
}

/*
	NewFilterFlow
	Create a Flow with a filter function.
	Have one input port and one output port.
	Only allow elements to go downstream.
*/
func NewFilterFlow(f func(interface{}) (bool, error), options ...queue.Option) Flow {
	in := queue.NewQueueEmpty(options...)
	out := queue.NewQueueEmpty(options...)
	return &flowImpl{
		in:  in,
		out: out,
		graphTree: MapGraph(in, out, func(i interface{}) (interface{}, error) {
			ok, err := f(i)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, PassPermissionError
			}
			return i, nil
		}),
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
