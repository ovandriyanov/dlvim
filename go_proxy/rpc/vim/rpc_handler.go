package vim

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	dlvapi "github.com/go-delve/delve/service/api"
	dlvrpc "github.com/go-delve/delve/service/rpc2"
	"github.com/ovandriyanov/dlvim/go_proxy/rpc/dlv"
	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
	_ "github.com/ovandriyanov/dlvim/go_proxy/upstream/command"
	"github.com/ovandriyanov/dlvim/go_proxy/vimevent"
	"golang.org/x/xerrors"
)

const (
	ServiceName = "Dlvim"
)

type RPCHandler struct {
	server *Server
	ctx    context.Context
}

func fqmn(method string) string {
	return fmt.Sprintf("%s.%s", ServiceName, method)
}

var KnownMethods map[string]struct{}

func init() {
	KnownMethods = make(map[string]struct{})
	handlerType := reflect.TypeOf((*RPCHandler)(nil))
	log.Println("Known methods:")
	for i := 0; i < handlerType.NumMethod(); i++ {
		method := handlerType.Method(i)
		knownMethodName := fqmn(method.Name)
		KnownMethods[knownMethodName] = struct{}{}
		log.Printf("    %s\n", knownMethodName)
	}
}

func (h *RPCHandler) Foo(req map[string]interface{}, resp *map[string]interface{}) error {
	(*resp)["foo"] = "bar"
	return nil
}

func (h *RPCHandler) Initialize(req map[string]interface{}, resp *map[string]interface{}) error {
	argvInterface, ok := req["dlv_argv"]
	if !ok {
		return xerrors.New("missing required parameter 'dlv_argv'")
	}

	dlvArgv, ok := argvInterface.([]interface{})
	if !ok {
		return xerrors.Errorf("expected 'dlv_argv' parameter to be string array but got %T", argvInterface)
	}

	var stringArgv []string
	for i, arg := range dlvArgv {
		strArg, ok := arg.(string)
		if !ok {
			return xerrors.Errorf("expected dlv_argv[%d] to be string but got %T", i, arg)
		}
		stringArgv = append(stringArgv, strArg)
	}

	command, err := upstream.NewCommand(stringArgv)
	if err != nil {
		return xerrors.Errorf("'dlv_argv' is invalid: %w", err)
	}

	inventory, err := h.server.Initialize(command)
	if err != nil {
		return err
	}
	(*resp)["proxy_listen_address"] = inventory.ProxyListenAddress().String()

	return nil
}

func (h *RPCHandler) ListBreakpoints(req map[string]interface{}, resp *map[string]interface{}) error {
	if h.server.inventory == nil {
		return xerrors.New("not initialized")
	}
	if err := h.server.inventory.upstreamClient.Call(dlv.FQMN("ListBreakpoints"), req, resp); err != nil {
		return err
	}
	clearDefaultBreakpoints(*resp)
	return nil
}

func clearDefaultBreakpoints(listBreakpointsResponse map[string]interface{}) {
	breakpointsTypeErased, ok := listBreakpointsResponse["Breakpoints"]
	if !ok {
		log.Println("WARNING: no \"Breakpoints\" key found in the dlv ListBreakpoints response")
		return
	}
	breakpoints, ok := breakpointsTypeErased.([]interface{})
	if !ok {
		log.Printf("WARNING: expected \"Breakpoints\" to be a list, not %T\n", breakpointsTypeErased)
		return
	}

	j := len(breakpoints)
	for i := 0; i < j; {
		breakpoint, ok := breakpoints[i].(map[string]interface{})
		if !ok {
			log.Printf("WARNING: expected Breakpoints[%d] to be a map[string]interface{}, not %T\n", i, breakpoints[i])
			i++
			continue
		}
		idTypeErased, ok := breakpoint["id"]
		if !ok {
			log.Printf("WARNING: Breakpoints[%d] has no field named \"id\"\n", i)
			i++
			continue
		}
		id, ok := idTypeErased.(float64) // TODO: use json.Number or something
		if !ok {
			log.Printf("WARNING: expected Breakpoints[%d].name to be an int, not %T\n", i, idTypeErased)
			i++
			continue
		}
		if id >= 0 {
			i++
			continue
		}
		breakpoints[i], breakpoints[j-1] = breakpoints[j-1], breakpoints[i]
		j--
	}
	listBreakpointsResponse["Breakpoints"] = breakpoints[:j]
}

func (h *RPCHandler) GetNextEvent(req map[string]interface{}, resp *map[string]interface{}) error {
	var event vimevent.Event
	select {
	case event = <-h.server.events:
	case <-h.ctx.Done():
		return context.Canceled
	}

	(*resp)["kind"] = event.Kind()
	(*resp)["payload"] = event
	return nil
}

type CreateOrDeleteBreakpointIn struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type CreateOrDeleteBreakpointOut struct{}

func jsonDump(anything interface{}) string {
	marshaledData, _ := json.MarshalIndent(anything, "", "    ")
	return string(marshaledData)
}

func (h *RPCHandler) CreateOrDeleteBreakpoint(req *CreateOrDeleteBreakpointIn, resp *CreateOrDeleteBreakpointOut) error {
	upstreamClient := h.server.UpstreamClient()
	if upstreamClient == nil {
		return xerrors.New("not initialized")
	}

	findLocationRequest := dlvrpc.FindLocationIn{
		Scope: dlvapi.EvalScope{GoroutineID: -1},
		Loc:   fmt.Sprintf("%s:%d", req.File, req.Line),
	}
	var findLocationResponse dlvrpc.FindLocationOut
	if err := upstreamClient.Call(dlv.FQMN("FindLocation"), &findLocationRequest, &findLocationResponse); err != nil {
		return xerrors.Errorf("cannot find location: %w", err)
	}
	if len(findLocationResponse.Locations) == 0 {
		return xerrors.New("no locations found")
	}
	foundLocation := findLocationResponse.Locations[0]

	listBreakpointsRequest := dlvrpc.ListBreakpointsIn{All: false}
	var listBreakpointsResponse dlvrpc.ListBreakpointsOut
	if err := upstreamClient.Call(dlv.FQMN("ListBreakpoints"), &listBreakpointsRequest, &listBreakpointsResponse); err != nil {
		return xerrors.Errorf("cannot list breakpoints: %w", err)
	}
	for _, breakpoint := range listBreakpointsResponse.Breakpoints {
		for _, addr := range breakpoint.Addrs {
			if addr == foundLocation.PC {
				clearBreakpointRequest := dlvrpc.ClearBreakpointIn{Id: breakpoint.ID}
				var clearBreakpointResponse dlvrpc.ClearBreakpointOut
				if err := upstreamClient.Call(dlv.FQMN("ClearBreakpoint"), &clearBreakpointRequest, &clearBreakpointResponse); err != nil {
					return xerrors.Errorf("cannot clear breakpoint %d: %w", breakpoint.ID, err)
				}
				log.Printf("ClearBreakpoint response: %v\n", jsonDump(clearBreakpointResponse))
				return nil
			}
		}
	}

	createBreakpointRequest := dlvrpc.CreateBreakpointIn{
		Breakpoint: dlvapi.Breakpoint{
			Addrs: []uint64{foundLocation.PC},
		},
	}
	var createBreakpointResponse dlvrpc.CreateBreakpointOut
	if err := upstreamClient.Call(dlv.FQMN("CreateBreakpoint"), &createBreakpointRequest, &createBreakpointResponse); err != nil {
		return xerrors.Errorf("cannot create breakpoint: %w", err)
	}

	log.Printf("CreateBreakpoint response: %v\n", jsonDump(createBreakpointResponse))
	return nil
}

type GetStateIn struct{}
type State struct {
	File string `json:"file"`
	Line int    `json:"line"`
}
type GetStateOut struct {
	State *State `json:"state"`
}

func (h *RPCHandler) GetState(req *GetStateIn, resp *GetStateOut) error {
	upstreamClient := h.server.UpstreamClient()
	if upstreamClient == nil {
		return xerrors.New("not initialized")
	}

	getStateRequest := dlvrpc.StateIn{NonBlocking: false}
	var getStateResponse dlvrpc.StateOut
	if err := upstreamClient.Call(dlv.FQMN("State"), &getStateRequest, &getStateResponse); err != nil {
		return err
	}

	state := getStateResponse.State
	if state == nil {
		return xerrors.New("upstream returned zero state")
	}
	selectedGoroutine := state.SelectedGoroutine
	if selectedGoroutine == nil {
		// This usually means the program finished it's execution
		return nil
	}

	resp.State = &State{
		File: selectedGoroutine.UserCurrentLoc.File,
		Line: selectedGoroutine.UserCurrentLoc.Line,
	}
	return nil
}

type ContinueIn struct{}
type ContinueOut struct{}

func (h *RPCHandler) Continue(req *ContinueIn, resp *ContinueOut) error {
	upstreamClient := h.server.UpstreamClient()
	if upstreamClient == nil {
		return xerrors.New("not initialized")
	}

	commandRequest := dlvapi.DebuggerCommand{Name: dlvapi.Continue}
	var commandResponse dlvrpc.CommandOut
	return upstreamClient.Call(dlv.FQMN("Command"), &commandRequest, &commandResponse)
}

type NextIn struct{}
type NextOut struct{}

func (h *RPCHandler) Next(req *NextIn, resp *NextOut) error {
	upstreamClient := h.server.UpstreamClient()
	if upstreamClient == nil {
		return xerrors.New("not initialized")
	}

	commandRequest := dlvapi.DebuggerCommand{Name: dlvapi.Next}
	var commandResponse dlvrpc.CommandOut
	return upstreamClient.Call(dlv.FQMN("Next"), &commandRequest, &commandResponse)
}

type StepIn struct{}
type StepOut struct{}

func (h *RPCHandler) Step(req *StepIn, resp *StepOut) error {
	upstreamClient := h.server.UpstreamClient()
	if upstreamClient == nil {
		return xerrors.New("not initialized")
	}

	commandRequest := dlvapi.DebuggerCommand{Name: dlvapi.Step}
	var commandResponse dlvrpc.CommandOut
	return upstreamClient.Call(dlv.FQMN("Step"), &commandRequest, &commandResponse)
}

type StepoutIn struct{}
type StepoutOut struct{}

func (h *RPCHandler) Stepout(req *StepoutIn, resp *StepoutOut) error {
	upstreamClient := h.server.UpstreamClient()
	if upstreamClient == nil {
		return xerrors.New("not initialized")
	}

	commandRequest := dlvapi.DebuggerCommand{Name: dlvapi.StepOut}
	var commandResponse dlvrpc.CommandOut
	return upstreamClient.Call(dlv.FQMN("Stepout"), &commandRequest, &commandResponse)
}

func NewRPCHandler(server *Server, ctx context.Context) *RPCHandler {
	return &RPCHandler{
		server: server,
		ctx:    ctx,
	}
}
