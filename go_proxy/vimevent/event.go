package vimevent

type Event interface {
	Kind() string
}

type BreakpointsUpdated struct{}
type StateUpdated struct{}

func (*BreakpointsUpdated) Kind() string {
	return "BREAKPOINTS_UPDATED"
}

func (*StateUpdated) Kind() string {
	return "STATE_UPDATED"
}
