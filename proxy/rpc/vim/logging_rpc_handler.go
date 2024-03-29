// Generated with generate_logging_rpc_handler. Do not edit!

package vim

import (
	json "encoding/json"
	log "log"
)

type LoggingRPCHandler struct {
	serverName string
	wrappedHandler *RPCHandler
}

func (h *LoggingRPCHandler) Foo(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Foo %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Foo(request, response)
	if err != nil {
		log.Printf("%s: --> Foo error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Foo %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Initialize(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Initialize %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Initialize(request, response)
	if err != nil {
		log.Printf("%s: --> Initialize error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Initialize %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) ListBreakpoints(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- ListBreakpoints %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.ListBreakpoints(request, response)
	if err != nil {
		log.Printf("%s: --> ListBreakpoints error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> ListBreakpoints %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) GetNextEvent(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- GetNextEvent %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.GetNextEvent(request, response)
	if err != nil {
		log.Printf("%s: --> GetNextEvent error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> GetNextEvent %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) CreateOrDeleteBreakpoint(request *CreateOrDeleteBreakpointIn, response *CreateOrDeleteBreakpointOut) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- CreateOrDeleteBreakpoint %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.CreateOrDeleteBreakpoint(request, response)
	if err != nil {
		log.Printf("%s: --> CreateOrDeleteBreakpoint error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> CreateOrDeleteBreakpoint %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Continue(request *ContinueIn, response *ContinueOut) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Continue %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Continue(request, response)
	if err != nil {
		log.Printf("%s: --> Continue error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Continue %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Next(request *NextIn, response *NextOut) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Next %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Next(request, response)
	if err != nil {
		log.Printf("%s: --> Next error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Next %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Step(request *StepIn, response *StepOut) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Step %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Step(request, response)
	if err != nil {
		log.Printf("%s: --> Step error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Step %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Stepout(request *StepoutIn, response *StepoutOut) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Stepout %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Stepout(request, response)
	if err != nil {
		log.Printf("%s: --> Stepout error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Stepout %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Up(request *UpIn, response *UpOut) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Up %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Up(request, response)
	if err != nil {
		log.Printf("%s: --> Up error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Up %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Down(request *DownIn, response *DownOut) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Down %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Down(request, response)
	if err != nil {
		log.Printf("%s: --> Down error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Down %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) SwitchStackFrame(request *SwitchStackFrameIn, response *SwitchStackFrameOut) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- SwitchStackFrame %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.SwitchStackFrame(request, response)
	if err != nil {
		log.Printf("%s: --> SwitchStackFrame error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> SwitchStackFrame %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Evaluate(request *EvaluateIn, response *EvaluateOut) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Evaluate %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Evaluate(request, response)
	if err != nil {
		log.Printf("%s: --> Evaluate error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Evaluate %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) ListGoroutines(request *ListGoroutinesIn, response *ListGoroutinesOut) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- ListGoroutines %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.ListGoroutines(request, response)
	if err != nil {
		log.Printf("%s: --> ListGoroutines error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> ListGoroutines %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) SwitchGoroutine(request *SwitchGoroutineIn, response *CommandOut) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- SwitchGoroutine %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.SwitchGoroutine(request, response)
	if err != nil {
		log.Printf("%s: --> SwitchGoroutine error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> SwitchGoroutine %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func NewLoggingRPCHandler(wrappedHandler *RPCHandler, serverName string) *LoggingRPCHandler {
	return &LoggingRPCHandler{
		serverName: serverName,
		wrappedHandler: wrappedHandler,
	}
}
