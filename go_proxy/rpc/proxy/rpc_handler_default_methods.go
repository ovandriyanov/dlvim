package proxy

func (h *RPCHandler) SetApiVersion(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("SetApiVersion"), req, resp)
}

func (h *RPCHandler) IsMulticlient(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("IsMulticlient"), req, resp)
}

func (h *RPCHandler) State(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("State"), req, resp)
}

func (h *RPCHandler) ListFunctions(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("ListFunctions"), req, resp)
}

func (h *RPCHandler) AttachedToExistingProcess(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("AttachedToExistingProcess"), req, resp)
}

func (h *RPCHandler) Detach(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("Detach"), req, resp)
}

func (h *RPCHandler) Recorded(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("Recorded"), req, resp)
}

func (h *RPCHandler) Command(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("Command"), req, resp)
}

func (h *RPCHandler) FindLocation(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("FindLocation"), req, resp)
}

func (h *RPCHandler) LastModified(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("LastModified"), req, resp)
}

func (h *RPCHandler) Stacktrace(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("Stacktrace"), req, resp)
}

func (h *RPCHandler) ProcessPid(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("Restart"), req, resp)
}

func (h *RPCHandler) Restart(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("Restart"), req, resp)
}

func (h *RPCHandler) Eval(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("Eval"), req, resp)
}

func (h *RPCHandler) GetBreakpoint(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("GetBreakpoint"), req, resp)
}

func (h *RPCHandler) ListBreakpoints(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("ListBreakpoints"), req, resp)
}

func (h *RPCHandler) ListGoroutines(req map[string]interface{}, resp *map[string]interface{}) error {
	return h.defaultHandler(fqmn("ListGoroutines"), req, resp)
}
