package gooastream

type (
	Flow interface {
		Via(Flow) Flow
		To(Sink) Sink

		Inlet
		Outlet
		Graph
	}

	flowImpl struct {
		in    Queue
		out   Queue
		graph *graph
	}
	//balancerFlowImpl struct {
	//	id     string
	//	in     Queue
	//	outs   []Queue
	//	graphs graphs
	//}
	//margeFlowImpl struct {
	//	id     string
	//	ins    []Queue
	//	out    Queue
	//	graphs graphs
	//}
)

//
//func BalanceOut(flow Flow, size int, buffer int) []Flow {
//	var gs graphs
//	var outs []Queue
//
//	var flows []Flow
//	for i := 0; i < size; i++ {
//		balanceOutFlow := NewBufferFlow(buffer)
//		flow.Graphs()
//		balanceOutFlow.Graphs()
//		flows = append(flows, balanceOutFlow)
//	}
//	return flows
//}
//func (a *balancerFlowImpl) Via(flows MultiInputFlow) Flow {
//	if len(a.getOuts()) != len(flows.getIns()) {
//		panic("in out port is different")
//	}
//
//	for i := range a.getOuts() {
//		a.graphs = append(a.graphs, &graph{
//			id:   fmt.Sprintf("balance flow via %d", i),
//			from: a.getOuts()[i],
//			to:   flows.getIns()[i],
//			task: emptyTaskFunc,
//		})
//		a.graphs = append(a.graphs, flows.Graphs()...)
//	}
//	return &flowImpl{
//		in:     a.In(),
//		out:    flows.Out(),
//		graphs: a.graphs,
//	}
//}
//func (a *balancerFlowImpl) IVia(flows []Flow) MultiOutputFlow {
//	if len(a.getOuts()) != len(flows) {
//		panic("in out port is different")
//	}
//
//	var outs []Queue
//	for i, v := range a.getOuts() {
//		flow := flows[i]
//		a.graphs = append(a.graphs, &graph{
//			id:   fmt.Sprintf("balance flow ivia %d", i),
//			from: v,
//			to:   flow.In(),
//			task: emptyTaskFunc,
//		})
//		a.graphs = append(a.graphs, flow.Graphs()...)
//		outs = append(outs, flow.Out())
//	}
//	return &balancerFlowImpl{
//		in:     a.In(),
//		outs:   outs,
//		graphs: a.graphs,
//	}
//}
//func (a *balancerFlowImpl) ITo(sinks []Sink) Sink {
//	if len(a.getOuts()) != len(sinks) {
//		panic("in out port is different")
//	}
//
//	for i := range a.getOuts() {
//		sink := sinks[i]
//		a.graphs = append(a.graphs, &graph{
//			id:   fmt.Sprintf("balance flow ito %d", i),
//			from: a.getOuts()[i],
//			to:   sink.In(),
//			task: emptyTaskFunc,
//		})
//		a.graphs = append(a.graphs, sink.Graphs()...)
//	}
//	return &sinkImpl{
//		in:     a.In(),
//		graphs: a.graphs,
//	}
//}
//func (a *balancerFlowImpl) In() Queue {
//	return a.in
//}
//func (a *balancerFlowImpl) getOuts() []Queue {
//	return a.outs
//}
//func (a *balancerFlowImpl) Graphs() graphs {
//	return a.graphs
//}
//
//func NewMargeFlow(size int, buffer int) MultiInputFlow {
//	out := NewQueueEmpty(buffer)
//	var gs graphs
//	var ins []Queue
//	for i := 0; i < size; i++ {
//		marge := NewQueueEmpty(buffer)
//		gs = append(gs, &graph{
//			id:   fmt.Sprintf("marge flow %d", i),
//			from: marge,
//			to:   out,
//			task: emptyTaskFunc,
//		})
//		ins = append(ins, marge)
//	}
//	return &margeFlowImpl{
//		id:     "marge flow",
//		ins:    ins,
//		out:    out,
//		graphs: gs,
//	}
//}
//func (a *margeFlowImpl) Via(flow Flow) MultiInputFlow {
//	a.graphs = append(a.graphs, &graph{
//		id:   "marge flow via",
//		from: a.out,
//		to:   flow.In(),
//		task: emptyTaskFunc,
//	})
//	a.graphs = append(a.graphs, flow.Graphs()...)
//	return &margeFlowImpl{
//		ins:    a.getIns(),
//		out:    flow.Out(),
//		graphs: a.graphs,
//	}
//}
//func (a *margeFlowImpl) getIns() []Queue {
//	return a.ins
//}
//func (a *margeFlowImpl) Out() Queue {
//	return a.out
//}
//func (a *margeFlowImpl) Graphs() graphs {
//	return a.graphs
//}

var _ Flow = (*flowImpl)(nil)

func NewBufferFlow(buffer int) Flow {
	in := NewQueueEmpty(buffer)
	out := NewQueueEmpty(buffer)
	return &flowImpl{
		in:    in,
		out:   out,
		graph: passThrowGraph(in, out),
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
		in:    sink.In(),
		out:   source.Out(),
		graph: sink.getGraph().Append(source.getGraph()),
	}
}

func (a *flowImpl) Via(flow Flow) Flow {
	return &flowImpl{
		in:  a.In(),
		out: flow.Out(),
		graph: a.graph.
			Append(passThrowGraph(a.Out(), flow.In())).
			Append(flow.getGraph()),
	}
}
func (a *flowImpl) To(sink Sink) Sink {
	return &sinkImpl{
		in: a.In(),
		graph: a.graph.
			Append(passThrowGraph(a.Out(), sink.In())).
			Append(sink.getGraph()),
	}
}
func (a *flowImpl) In() Queue {
	return a.in
}
func (a *flowImpl) Out() Queue {
	return a.out
}
func (a *flowImpl) getGraph() *graph {
	return a.graph
}
