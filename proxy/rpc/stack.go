package rpc

import (
	"github.com/go-delve/delve/service/api"
)

type StackFrame struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

func NewStackTrace(stackFrames []api.Stackframe) (trace []StackFrame) {
	for _, location := range stackFrames {
		trace = append(trace, StackFrame{File: location.File, Line: location.Line})
	}
	return trace
}
