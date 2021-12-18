package proxy

import (
	"github.com/ovandriyanov/dlvim/go_proxy/rpc/dlv"
)

func (h *RPCHandler) SetApiVersion(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("SetApiVersion"), req, resp)
}

func (h *RPCHandler) IsMulticlient(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("IsMulticlient"), req, resp)
}

func (h *RPCHandler) State(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("State"), req, resp)
}

func (h *RPCHandler) ListFunctions(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("ListFunctions"), req, resp)
}

func (h *RPCHandler) AttachedToExistingProcess(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("AttachedToExistingProcess"), req, resp)
}

func (h *RPCHandler) Detach(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("Detach"), req, resp)
}

func (h *RPCHandler) Recorded(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("Recorded"), req, resp)
}

func (h *RPCHandler) Command(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("Command"), req, resp)
}

func (h *RPCHandler) FindLocation(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("FindLocation"), req, resp)
}

func (h *RPCHandler) LastModified(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("LastModified"), req, resp)
}

func (h *RPCHandler) Stacktrace(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("Stacktrace"), req, resp)
}

func (h *RPCHandler) ProcessPid(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("Restart"), req, resp)
}

func (h *RPCHandler) Restart(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("Restart"), req, resp)
}

func (h *RPCHandler) Eval(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("Eval"), req, resp)
}

func (h *RPCHandler) GetBreakpoint(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("GetBreakpoint"), req, resp)
}

func (h *RPCHandler) ListBreakpoints(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("ListBreakpoints"), req, resp)
}

func (h *RPCHandler) ListGoroutines(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(dlv.FQMN("ListGoroutines"), req, resp)
}
