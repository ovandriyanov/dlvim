package vimevent

type Event interface {
	Kind() string
}

type BreakpointsUpdated struct{}
type CommandIssued struct{}
type StateUpdated struct{}

func (*BreakpointsUpdated) Kind() string {
	return "BREAKPOINTS_UPDATED"
}

func (*CommandIssued) Kind() string {
	return "COMMAND_ISSUED"
}

func (*StateUpdated) Kind() string {
	return "STATE_UPDATED"
}
