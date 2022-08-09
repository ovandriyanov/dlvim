package vim

import (
	"golang.org/x/xerrors"
)

type Stack struct {
	stackTrace   []StackFrame
	currentFrame int
}

func (s *Stack) IsTopmostFrame() bool {
	return s.currentFrame == 0
}

func (s *Stack) SetStackTrace(stackTrace []StackFrame) {
	s.stackTrace = stackTrace
	s.currentFrame = 0
}

func (s *Stack) Up() error {
	newCurrentFrame := s.currentFrame + 1
	if newCurrentFrame >= len(s.stackTrace) {
		return xerrors.New("cannot go up from the bottommost frame")
	}
	s.currentFrame = newCurrentFrame
	return nil
}

func (s *Stack) Down() error {
	newCurrentFrame := s.currentFrame - 1
	if newCurrentFrame < 0 {
		return xerrors.New("cannot go down from the topmost frame")
	}
	s.currentFrame = newCurrentFrame
	return nil
}

func (s *Stack) Trace() []StackFrame {
	return s.stackTrace
}

func (s *Stack) CurrentFrame() int {
	return s.currentFrame
}

func NewStack() *Stack {
	return &Stack{
		stackTrace:   nil,
		currentFrame: 0,
	}
}
