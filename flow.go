package gooastream

import (
	"fmt"
)

type (
	Flow interface {
		Via(Flow) Flow
		To(Sink) Sink

		SingleInputToSingleOutput
		Graph
	}
	MultiInputFlow interface {
		Via(Flow) MultiInputFlow

		MultiInputToSingleOutput
		Graph
	}
	MultiOutputFlow interface {
		Via(MultiInputFlow) Flow
		IVia([]Flow) MultiOutputFlow
		ITo([]Sink) Sink

		SingleInputToMultiOutput
		Graph
	}

	flowImpl struct {
		id     string
		in     Queue
		out    Queue
		graphs graphs
	}
	balancerFlowImpl struct {
		id     string
		in     Queue
		outs   []Queue
		graphs graphs
	}
	margeFlowImpl struct {
		id     string
		ins    []Queue
		out    Queue
		graphs graphs
	}
)

func NewBalanceFlow(size int, buffer int) MultiOutputFlow {
	in := NewQueueEmpty(buffer)
	var gs graphs
	var outs []Queue
	for i := 0; i < size; i++ {
		balancer := NewQueueEmpty(buffer)
		gs = append(gs, &graph{
			id:   fmt.Sprintf("balance flow %d", i),
			from: in,
			to:   balancer,
			task: func(v interface{}) interface{} {
				return v
			},
		})
		outs = append(outs, balancer)
	}
	return &balancerFlowImpl{
		id:     "balance flow",
		in:     in,
		outs:   outs,
		graphs: gs,
	}
}
func (a *balancerFlowImpl) Via(flows MultiInputFlow) Flow {
	if len(a.getOuts()) != len(flows.getIns()) {
		panic("in out port is different")
	}

	for i := range a.getOuts() {
		a.graphs = append(a.graphs, &graph{
			id:   fmt.Sprintf("balance flow via %d", i),
			from: a.getOuts()[i],
			to:   flows.getIns()[i],
			task: func(i interface{}) interface{} {
				return i
			},
		})
		a.graphs = append(a.graphs, flows.getGraphs()...)
	}
	return &flowImpl{
		in:     a.getIn(),
		out:    flows.getOut(),
		graphs: a.graphs,
	}
}
func (a *balancerFlowImpl) IVia(flows []Flow) MultiOutputFlow {
	if len(a.getOuts()) != len(flows) {
		panic("in out port is different")
	}

	var outs []Queue
	for i, v := range a.getOuts() {
		flow := flows[i]
		a.graphs = append(a.graphs, &graph{
			id:   fmt.Sprintf("balance flow ivia %d", i),
			from: v,
			to:   flow.getIn(),
			task: func(i interface{}) interface{} {
				return i
			},
		})
		a.graphs = append(a.graphs, flow.getGraphs()...)
		outs = append(outs, flow.getOut())
	}
	return &balancerFlowImpl{
		in:     a.getIn(),
		outs:   outs,
		graphs: a.graphs,
	}
}
func (a *balancerFlowImpl) ITo(sinks []Sink) Sink {
	if len(a.getOuts()) != len(sinks) {
		panic("in out port is different")
	}

	for i := range a.getOuts() {
		sink := sinks[i]
		a.graphs = append(a.graphs, &graph{
			id:   fmt.Sprintf("balance flow ito %d", i),
			from: a.getOuts()[i],
			to:   sink.getIn(),
			task: func(v interface{}) interface{} {
				return v
			},
		})
		a.graphs = append(a.graphs, sink.getGraphs()...)
	}
	return &sinkImpl{
		in:     a.getIn(),
		graphs: a.graphs,
	}
}
func (a *balancerFlowImpl) getIn() Queue {
	return a.in
}
func (a *balancerFlowImpl) getOuts() []Queue {
	return a.outs
}
func (a *balancerFlowImpl) getGraphs() graphs {
	return a.graphs
}

func NewMargeFlow(size int, buffer int) MultiInputFlow {
	out := NewQueueEmpty(buffer)
	var gs graphs
	var ins []Queue
	for i := 0; i < size; i++ {
		marge := NewQueueEmpty(buffer)
		gs = append(gs, &graph{
			id:   fmt.Sprintf("marge flow %d", i),
			from: marge,
			to:   out,
			task: func(i interface{}) interface{} {
				return i
			},
		})
		ins = append(ins, marge)
	}
	return &margeFlowImpl{
		id:     "marge flow",
		ins:    ins,
		out:    out,
		graphs: gs,
	}
}
func (a *margeFlowImpl) Via(flow Flow) MultiInputFlow {
	a.graphs = append(a.graphs, &graph{
		id:   "marge flow via",
		from: a.out,
		to:   flow.getIn(),
		task: func(i interface{}) interface{} {
			return i
		},
	})
	a.graphs = append(a.graphs, flow.getGraphs()...)
	return &margeFlowImpl{
		ins:    a.getIns(),
		out:    flow.getOut(),
		graphs: a.graphs,
	}
}
func (a *margeFlowImpl) getIns() []Queue {
	return a.ins
}
func (a *margeFlowImpl) getOut() Queue {
	return a.out
}
func (a *margeFlowImpl) getGraphs() graphs {
	return a.graphs
}

func NewFlow(task func(interface{}) interface{}, buffer int) Flow {
	in := NewQueueEmpty(buffer)
	out := NewQueueEmpty(buffer)
	return &flowImpl{
		id:  "flow",
		in:  in,
		out: out,
		graphs: graphs{
			&graph{
				id:   "flow",
				from: in,
				to:   out,
				task: task,
			},
		},
	}
}
func (a *flowImpl) Via(flow Flow) Flow {
	a.graphs = append(a.graphs, &graph{
		id:   "flow via",
		from: a.getOut(),
		to:   flow.getIn(),
		task: func(i interface{}) interface{} {
			return i
		},
	})
	a.graphs = append(a.graphs, flow.getGraphs()...)
	return &flowImpl{
		in:     a.getIn(),
		out:    flow.getOut(),
		graphs: a.graphs,
	}
}
func (a *flowImpl) To(sink Sink) Sink {
	a.graphs = append(a.graphs, &graph{
		id:   "flow to",
		from: a.getOut(),
		to:   sink.getIn(),
		task: func(i interface{}) interface{} {
			return i
		},
	})
	a.graphs = append(a.graphs, sink.getGraphs()...)
	return &sinkImpl{
		in:     a.getIn(),
		graphs: a.graphs,
	}
}
func (a *flowImpl) getIn() Queue {
	return a.in
}
func (a *flowImpl) getOut() Queue {
	return a.out
}
func (a *flowImpl) getGraphs() graphs {
	return a.graphs
}
