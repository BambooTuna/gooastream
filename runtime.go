package gooastream

type (
	Inlet interface {
		In() Queue
	}
	Outlet interface {
		Out() Queue
	}
)
