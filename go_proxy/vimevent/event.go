package vimevent

type Event interface {
	Kind() string
}

type BreakpointsUpdated struct{}

func (*BreakpointsUpdated) Kind() string {
	return "BREAKPOINTS_UPDATED"
}
