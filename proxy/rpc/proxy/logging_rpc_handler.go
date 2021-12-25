// Generated with generate_logging_rpc_handler. Do not edit!

package proxy

import (
	"encoding/json"
	"log"
)

type LoggingRPCHandler struct {
	serverName string
	wrappedHandler *RPCHandler
}

func (h *LoggingRPCHandler) CreateBreakpoint(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- CreateBreakpoint %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.CreateBreakpoint(request, response)
	if err != nil {
		log.Printf("%s: --> CreateBreakpoint error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> CreateBreakpoint %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) AmendBreakpoint(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- AmendBreakpoint %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.AmendBreakpoint(request, response)
	if err != nil {
		log.Printf("%s: --> AmendBreakpoint error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> AmendBreakpoint %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) ClearBreakpoint(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- ClearBreakpoint %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.ClearBreakpoint(request, response)
	if err != nil {
		log.Printf("%s: --> ClearBreakpoint error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> ClearBreakpoint %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Command(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Command %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Command(request, response)
	if err != nil {
		log.Printf("%s: --> Command error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Command %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) SetApiVersion(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- SetApiVersion %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.SetApiVersion(request, response)
	if err != nil {
		log.Printf("%s: --> SetApiVersion error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> SetApiVersion %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) IsMulticlient(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- IsMulticlient %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.IsMulticlient(request, response)
	if err != nil {
		log.Printf("%s: --> IsMulticlient error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> IsMulticlient %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) State(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- State %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.State(request, response)
	if err != nil {
		log.Printf("%s: --> State error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> State %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) ListFunctions(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- ListFunctions %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.ListFunctions(request, response)
	if err != nil {
		log.Printf("%s: --> ListFunctions error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> ListFunctions %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) AttachedToExistingProcess(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- AttachedToExistingProcess %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.AttachedToExistingProcess(request, response)
	if err != nil {
		log.Printf("%s: --> AttachedToExistingProcess error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> AttachedToExistingProcess %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Detach(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Detach %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Detach(request, response)
	if err != nil {
		log.Printf("%s: --> Detach error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Detach %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Recorded(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Recorded %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Recorded(request, response)
	if err != nil {
		log.Printf("%s: --> Recorded error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Recorded %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) FindLocation(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- FindLocation %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.FindLocation(request, response)
	if err != nil {
		log.Printf("%s: --> FindLocation error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> FindLocation %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) LastModified(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- LastModified %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.LastModified(request, response)
	if err != nil {
		log.Printf("%s: --> LastModified error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> LastModified %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Stacktrace(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Stacktrace %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Stacktrace(request, response)
	if err != nil {
		log.Printf("%s: --> Stacktrace error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Stacktrace %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) ProcessPid(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- ProcessPid %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.ProcessPid(request, response)
	if err != nil {
		log.Printf("%s: --> ProcessPid error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> ProcessPid %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Restart(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Restart %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Restart(request, response)
	if err != nil {
		log.Printf("%s: --> Restart error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Restart %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) Eval(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- Eval %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.Eval(request, response)
	if err != nil {
		log.Printf("%s: --> Eval error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> Eval %s\n", h.serverName, string(marshaledResponse))
	return nil
}

func (h *LoggingRPCHandler) GetBreakpoint(request map[string]interface{}, response *map[string]interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- GetBreakpoint %s\n", h.serverName, string(marshaledRequest))
	err := h.wrappedHandler.GetBreakpoint(request, response)
	if err != nil {
		log.Printf("%s: --> GetBreakpoint error %v\n", h.serverName, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> GetBreakpoint %s\n", h.serverName, string(marshaledResponse))
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

func (h *LoggingRPCHandler) ListGoroutines(request map[string]interface{}, response *map[string]interface{}) error {
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

func NewLoggingRPCHandler(wrappedHandler *RPCHandler, serverName string) *LoggingRPCHandler {
	return &LoggingRPCHandler{
		serverName: serverName,
		wrappedHandler: wrappedHandler,
	}
}
