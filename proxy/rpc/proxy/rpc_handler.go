//go:generate go run ../../../generators/generate_logging_rpc_handler --input-dir . --rpc-handler-type-name RPCHandler --output-file logging_rpc_handler.go
package proxy

import (
	"context"
	"log"
	"net/rpc"
	"reflect"

	dlvrpc "github.com/go-delve/delve/service/rpc2"
	commonrpc "github.com/ovandriyanov/dlvim/proxy/rpc"
	dlvimrpc "github.com/ovandriyanov/dlvim/proxy/rpc"
	"github.com/ovandriyanov/dlvim/proxy/rpc/dlv"
	"github.com/ovandriyanov/dlvim/proxy/vimevent"
)

var KnownMethods map[string]struct{}

func init() {
	KnownMethods = make(map[string]struct{})
	handlerType := reflect.TypeOf((*RPCHandler)(nil))
	log.Println("Known methods:")
	for i := 0; i < handlerType.NumMethod(); i++ {
		method := handlerType.Method(i)
		knownMethodName := dlv.FQMN(method.Name)
		KnownMethods[knownMethodName] = struct{}{}
		log.Printf("    %s\n", knownMethodName)
	}
}

type RPCHandler struct {
	ctx              context.Context
	dlvClient        *rpc.Client
	events           chan<- vimevent.Event
	handleStackTrace func([]commonrpc.StackFrame)
}

func (h *RPCHandler) defaultHandler(method string, req interface{}, resp interface{}) error {
	return h.dlvClient.Call(method, req, resp)
}

func (h *RPCHandler) CreateBreakpoint(req map[string]interface{}, resp *map[string]interface{}) error {
	err := h.defaultHandler(dlv.FQMN("CreateBreakpoint"), req, resp)
	if err != nil {
		return err
	}
	select {
	case h.events <- &vimevent.BreakpointsUpdated{}:
	case <-h.ctx.Done():
	}
	return nil
}

func (h *RPCHandler) AmendBreakpoint(req map[string]interface{}, resp *map[string]interface{}) error {
	err := h.defaultHandler(dlv.FQMN("AmendBreakpoint"), req, resp)
	if err != nil {
		return err
	}
	select {
	case h.events <- &vimevent.BreakpointsUpdated{}:
	case <-h.ctx.Done():
	}
	return nil
}

func (h *RPCHandler) ClearBreakpoint(req map[string]interface{}, resp *map[string]interface{}) error {
	err := h.defaultHandler(dlv.FQMN("ClearBreakpoint"), req, resp)
	if err != nil {
		return err
	}
	select {
	case h.events <- &vimevent.BreakpointsUpdated{}:
	case <-h.ctx.Done():
	}
	return nil

}

func NewRPCHandler(dlvClient *rpc.Client, events chan<- vimevent.Event, ctx context.Context, handleStackTrace func([]commonrpc.StackFrame)) *RPCHandler {
	return &RPCHandler{
		dlvClient:        dlvClient,
		events:           events,
		ctx:              ctx,
		handleStackTrace: handleStackTrace,
	}
}

func (h *RPCHandler) Command(req map[string]interface{}, resp *dlvrpc.CommandOut) error {
	select {
	case h.events <- &vimevent.CommandIssued{}:
	case <-h.ctx.Done():
	}

	err := h.defaultHandler(dlv.FQMN("Command"), req, resp)
	if err != nil {
		return err
	}

	var stackTraceResponse dlvrpc.StacktraceOut
	err = h.dlvClient.Call(dlv.FQMN("Stacktrace"), dlvrpc.StacktraceIn{Id: -1, Depth: 50}, &stackTraceResponse)
	if err != nil {
		log.Printf("WARNING: cannot get current stack trace: %s\n", err.Error())
	}
	h.handleStackTrace(commonrpc.NewStackTrace(stackTraceResponse.Locations))

	select {
	case h.events <- &vimevent.StateUpdated{
		State:      &resp.State,
		StackTrace: dlvimrpc.NewStackTrace(stackTraceResponse.Locations),
	}:
	case <-h.ctx.Done():
	}
	return nil
}
