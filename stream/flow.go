package stream

import (
	"github.com/BambooTuna/gooastream/queue"
)

type (
	flowImpl struct {
		in        queue.Queue
		out       queue.Queue
		graphTree *GraphTree
	}
)

var _ Flow = (*flowImpl)(nil)

func BuildFlow(in, out queue.Queue, graphTree *GraphTree) Flow {
	return &flowImpl{
		in:        in,
		out:       out,
		graphTree: graphTree,
	}
}

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
     +-------------------------------------------------------+
     | Resulting Flow                                       |
     |                                                      |
     |  +-------------+                  +---------------+  |
     |  |             |                  |               |  |
 I  ~~> | Sink        | [no-connection!] | Source        | ~~> O
     |  |             |                  |               |  |
     |  +-------------+                  +---------------+  |
     +------------------------------------------------------+

どちらかがクローズしても、その事はもう一方へ伝達されない
*/
func NewFlowFromSinkAndSource(sink Sink, source Source) Flow {
	return &flowImpl{
		in:  sink.In(),
		out: source.Out(),
		graphTree: sink.GraphTree().
			Append(source.GraphTree()),
	}
}

func NewFlowFanout(flow Flow, size int) []Flow {
	flows := make([]Flow, size)
	for i := 0; i < size; i++ {
		out := NewBufferFlow(0)
		flows[i] = &flowImpl{
			in:  flow.In(),
			out: out.Out(),
			graphTree: flow.GraphTree().
				Append(PassThrowGraph(flow.Out(), out.In())).
				Append(out.GraphTree()),
		}
	}
	return flows
}

func (a *flowImpl) Via(flow Flow) Flow {
	return &flowImpl{
		in:  a.In(),
		out: flow.Out(),
		graphTree: a.GraphTree().
			Append(PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	}
}
func (a *flowImpl) To(sink Sink) Sink {
	return &sinkImpl{
		in: a.In(),
		graphTree: a.GraphTree().
			Append(PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	}
}
func (a *flowImpl) In() queue.Queue {
	return a.in
}
func (a *flowImpl) Out() queue.Queue {
	return a.out
}
func (a *flowImpl) GraphTree() *GraphTree {
	return a.graphTree
}
