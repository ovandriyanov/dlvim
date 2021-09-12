package vim

const ServiceName = "Dlvim"

type RPCHandler struct{}

func NewRPCHandler() *RPCHandler {
	return &RPCHandler{}
}

func (h *RPCHandler) Foo(req map[string]interface{}, resp *map[string]interface{}) error {
	(*resp)["foo"] = "bar"
	return nil
}
