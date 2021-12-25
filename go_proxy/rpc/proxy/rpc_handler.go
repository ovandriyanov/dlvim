package proxy

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/rpc"
	"reflect"

	"github.com/ovandriyanov/dlvim/go_proxy/rpc/dlv"
	"github.com/ovandriyanov/dlvim/go_proxy/vimevent"
)

var KnownMethods map[string]struct{}

var debugRPC = flag.Bool("debug-rpc", false, "Show full requests and responses sent between dlv and the proxy")

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
	ctx       context.Context
	dlvClient *rpc.Client
	events    chan<- vimevent.Event
}

func (h *RPCHandler) defaultHandler(method string, req map[string]interface{}, resp *map[string]interface{}) error {
	if *debugRPC {
		jsonReq, _ := json.MarshalIndent(req, "", "    ")
		log.Printf("Call %s, request: %s\n", method, jsonReq)
	}
	err := h.dlvClient.Call(method, req, resp)
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
	if *debugRPC {
		jsonResp, _ := json.MarshalIndent(resp, "", "    ")
		log.Printf("Call %s, response: %s\n", method, jsonResp)
	}
	return err
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

func NewRPCHandler(dlvClient *rpc.Client, events chan<- vimevent.Event, ctx context.Context) *RPCHandler {
	return &RPCHandler{
		dlvClient: dlvClient,
		events:    events,
		ctx:       ctx,
	}
}

func (h *RPCHandler) Command(req map[string]interface{}, resp *map[string]interface{}) error {
	select {
	case h.events <- &vimevent.CommandIssued{}:
	case <-h.ctx.Done():
	}

	err := h.defaultHandler(dlv.FQMN("Command"), req, resp)
	if err != nil {
		return err
	}

	select {
	case h.events <- &vimevent.StateUpdated{}:
	case <-h.ctx.Done():
	}
	return nil
}
