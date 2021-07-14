package gooastream

type (
	SingleInput interface {
		getIn() Queue
	}
	SingleOutput interface {
		getOut() Queue
	}

	MultiInput interface {
		getIns() []Queue
	}
	MultiOutput interface {
		getOuts() []Queue
	}

	LeftOutput interface {
		getLeft() interface{}
	}
	RightOutput interface {
		getRight() interface{}
	}

	CloseToSingleOutput interface {
		SingleOutput
		LeftOutput
	}
	CloseToMultiOutput interface {
		MultiOutput
		LeftOutput
	}

	SingleInputToSingleOutput interface {
		SingleInput
		SingleOutput
	}
	SingleInputToMultiOutput interface {
		SingleInput
		MultiOutput
	}
	SingleInputToClose interface {
		SingleInput
		RightOutput
	}

	MultiInputToSingleOutput interface {
		MultiInput
		SingleOutput
	}
	MultiInputToMultiOutput interface {
		MultiInput
		MultiOutput
	}
	MultiInputToClose interface {
		MultiInput
		RightOutput
	}
)
