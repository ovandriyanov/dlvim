package main

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
