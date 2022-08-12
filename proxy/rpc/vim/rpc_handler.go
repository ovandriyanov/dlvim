//go:generate go run ../../../generators/generate_logging_rpc_handler --input-dir . --rpc-handler-type-name RPCHandler --output-file logging_rpc_handler.go
package vim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"

	dlvapi "github.com/go-delve/delve/service/api"
	dlvrpc "github.com/go-delve/delve/service/rpc2"
	"github.com/ovandriyanov/dlvim/proxy/rpc/dlv"
	"github.com/ovandriyanov/dlvim/proxy/upstream"
	"github.com/ovandriyanov/dlvim/proxy/vimevent"
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

	command, err := upstream.NewStartOption(stringArgv)
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

	return nil
}

type StackFrame struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type CommandOut struct {
	State      *dlvapi.DebuggerState `json:"state"`
	StackTrace []StackFrame          `json:"stack_trace"`
}

type ContinueIn struct{}
type ContinueOut CommandOut

func (h *RPCHandler) command(command string, response *CommandOut, goroutineID int) error {
	if command != "switchGoroutine" && !h.server.inventory.stack.IsTopmostFrame() {
		return xerrors.New("not on the topmost frame")
	}

	upstreamClient := h.server.UpstreamClient()
	if upstreamClient == nil {
		return xerrors.New("not initialized")
	}

	commandRequest := dlvapi.DebuggerCommand{
		Name:        command,
		GoroutineID: goroutineID,
	}
	var commandResponse dlvrpc.CommandOut
	err := upstreamClient.Call(dlv.FQMN("Command"), &commandRequest, &commandResponse)
	if err != nil {
		return err
	}
	response.State = &commandResponse.State

	var stackTraceResponse dlvrpc.StacktraceOut
	err = upstreamClient.Call(dlv.FQMN("Stacktrace"), dlvrpc.StacktraceIn{Id: -1, Depth: 50}, &stackTraceResponse)
	if err != nil {
		log.Printf("ERROR: cannot get stack trace: %s\n", err)
		return nil
	}
	for _, location := range stackTraceResponse.Locations {
		response.StackTrace = append(response.StackTrace, StackFrame{File: location.File, Line: location.Line})
	}
	h.server.inventory.stack.SetStackTrace(response.StackTrace)
	return nil
}

func (h *RPCHandler) Continue(req *ContinueIn, resp *ContinueOut) error {
	return h.command(dlvapi.Continue, (*CommandOut)(resp), 0)
}

type NextIn struct{}
type NextOut CommandOut

func (h *RPCHandler) Next(req *NextIn, resp *NextOut) error {
	return h.command(dlvapi.Next, (*CommandOut)(resp), 0)
}

type StepIn struct{}
type StepOut CommandOut

func (h *RPCHandler) Step(req *StepIn, resp *StepOut) error {
	return h.command(dlvapi.Step, (*CommandOut)(resp), 0)
}

type StepoutIn struct{}
type StepoutOut CommandOut

func (h *RPCHandler) Stepout(req *StepoutIn, resp *StepoutOut) error {
	return h.command(dlvapi.StepOut, (*CommandOut)(resp), 0)
}

type StackFrameOut struct {
	StackTrace   []StackFrame `json:"stack_trace"`
	CurrentFrame int          `json:"current_stack_frame"`
}

type UpIn struct{}
type UpOut StackFrameOut

func (h *RPCHandler) Up(req *UpIn, resp *UpOut) error {
	if err := h.server.inventory.stack.Up(); err != nil {
		return err
	}
	*resp = UpOut{
		StackTrace:   h.server.inventory.stack.Trace(),
		CurrentFrame: h.server.inventory.stack.CurrentFrame(),
	}
	return nil
}

type DownIn struct{}
type DownOut StackFrameOut

func (h *RPCHandler) Down(req *DownIn, resp *DownOut) error {
	if err := h.server.inventory.stack.Down(); err != nil {
		return err
	}
	*resp = DownOut{
		StackTrace:   h.server.inventory.stack.Trace(),
		CurrentFrame: h.server.inventory.stack.CurrentFrame(),
	}
	return nil
}

type SwitchStackFrameIn struct {
	StackFrame int `json:"stack_frame"`
}
type SwitchStackFrameOut StackFrameOut

func (h *RPCHandler) SwitchStackFrame(req *SwitchStackFrameIn, resp *SwitchStackFrameOut) error {
	if err := h.server.inventory.stack.SwitchFrame(req.StackFrame); err != nil {
		return err
	}
	*resp = SwitchStackFrameOut{
		StackTrace:   h.server.inventory.stack.Trace(),
		CurrentFrame: h.server.inventory.stack.CurrentFrame(),
	}
	return nil
}

type EvaluateIn struct {
	// Specify either line + cursor_position or expression
	Line           string `json:"line"`
	CursorPosition int    `json:"cursor_position"`

	Expression string `json:"expression"`
}

type EvaluateOut struct {
	OneLine string   `json:"one_line"`
	Pretty  []string `json:"pretty"`
}

func parseEvaluateExprRequest(request *EvaluateIn) (string, error) {
	if request.CursorPosition >= len(request.Line) || request.CursorPosition < 0 {
		return "", xerrors.Errorf("cursor position is outside range [0, %d)", len(request.Line))
	}
	forwardRegexp := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	backwardRegexp := regexp.MustCompile(`[^a-zA-Z0-9_.]`)

	locations := forwardRegexp.FindStringIndex(request.Line[request.CursorPosition:])
	end := len(request.Line)
	if locations != nil {
		end = locations[0] + request.CursorPosition
	}

	reversedInput := make([]byte, 0, request.CursorPosition)
	for i := request.CursorPosition; i >= 0; i-- {
		reversedInput = append(reversedInput, request.Line[i])
	}
	locations = backwardRegexp.FindIndex(reversedInput)
	begin := 0
	if locations != nil {
		begin = request.CursorPosition - locations[0] + 1
	}
	if begin > request.CursorPosition {
		begin = request.CursorPosition
	}

	return request.Line[begin:end], nil
}

func (h *RPCHandler) Evaluate(req *EvaluateIn, resp *EvaluateOut) error {
	var expr string
	if req.Expression != "" {
		expr = req.Expression
	} else {
		var err error
		expr, err = parseEvaluateExprRequest(req)
		if err != nil {
			return xerrors.Errorf("cannot parse expression: %w", err)
		}
	}

	var upstreamResp dlvrpc.EvalOut
	err := h.server.inventory.upstreamClient.Call(
		dlv.FQMN("Eval"),
		dlvrpc.EvalIn{
			Scope: dlvapi.EvalScope{
				GoroutineID:  -1,
				Frame:        h.server.inventory.stack.CurrentFrame(),
				DeferredCall: 0,
			},
			Expr: expr,
			Cfg: &dlvapi.LoadConfig{
				FollowPointers:     true,
				MaxVariableRecurse: 1,
				MaxStringLen:       10_000,
				MaxArrayValues:     100,
				MaxStructFields:    -1,
			},
		},
		&upstreamResp,
	)
	if err != nil {
		return err
	}
	*resp = EvaluateOut{
		OneLine: upstreamResp.Variable.Value,
		Pretty:  formatVariable(upstreamResp.Variable),
	}
	return nil
}

type ListGoroutinesIn struct {
	CurrentGoroutineID int `json:"current_goroutine_id"`
}

type ListGoroutinesOut struct {
	Goroutines            []string `json:"goroutines"`
	CurrentGoroutineIndex int      `json:"current_goroutine_index"`
}

func (h *RPCHandler) ListGoroutines(req *ListGoroutinesIn, resp *ListGoroutinesOut) error {
	var upstreamResp dlvrpc.ListGoroutinesOut
	if err := h.server.inventory.upstreamClient.Call(dlv.FQMN("ListGoroutines"), &dlvrpc.ListGoroutinesIn{}, &upstreamResp); err != nil {
		return err
	}
	buffer := bytes.NewBuffer(nil)
	tabWriter := tabwriter.NewWriter(buffer, 0, 1, 1, ' ', 0)
	currentGoroutineIndex := -1
	for i, goroutine := range upstreamResp.Goroutines {
		fmt.Fprintf(
			tabWriter,
			"%d\t%s\t%s:%d\n",
			goroutine.ID,
			goroutine.CurrentLoc.Function.Name(),
			goroutine.CurrentLoc.File,
			goroutine.CurrentLoc.Line,
		)
		if goroutine.ID == req.CurrentGoroutineID {
			currentGoroutineIndex = i
		}
	}
	_ = tabWriter.Flush()
	lines := strings.Split(buffer.String(), "\n")
	if len(lines) > 0 {
		lines = lines[:len(lines)-1] // the last line is always empty
	}
	*resp = ListGoroutinesOut{
		Goroutines:            lines,
		CurrentGoroutineIndex: currentGoroutineIndex,
	}
	return nil
}

type SwitchGoroutineIn struct {
	Line string `json:"line"`
}

type SwitchGoroutineOut struct{}

func (h *RPCHandler) SwitchGoroutine(req *SwitchGoroutineIn, resp *CommandOut) error {
	goroutineStringID := strings.Split(req.Line, " ")[0]
	goroutineID, err := strconv.Atoi(goroutineStringID)
	if err != nil {
		return xerrors.Errorf("invalid goroutine ID: %q", goroutineStringID)
	}
	return h.command("switchGoroutine", resp, goroutineID)
}

func formatVariable(variable *dlvapi.Variable) []string {
	object := toObject(variable)
	marshaled, _ := json.MarshalIndent(object, "", "  ")
	return strings.Split(string(marshaled), "\n")
}

// Converts a variable to an object ready to be marshaled into a JSON
func toObject(variable *dlvapi.Variable) (object interface{}) {
	if len(variable.Children) == 0 {
		return variable.Value
	}
	if strings.HasPrefix(variable.Type, "[]") {
		list := make([]interface{}, 0, len(variable.Children))
		for _, child := range variable.Children {
			list = append(list, toObject(&child))
		}
		object = list
	} else {
		dict := make(map[string]interface{})
		for _, child := range variable.Children {
			dict[child.Name] = toObject(&child)
		}
		object = dict
	}
	return object
}

func NewRPCHandler(server *Server, ctx context.Context) *RPCHandler {
	return &RPCHandler{
		server: server,
		ctx:    ctx,
	}
}
