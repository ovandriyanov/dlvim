package vim

const ServiceName = "Dlvim"

type RPCHandler struct{}

func NewRPCHandler() *RPCHandler {
	return &RPCHandler{}
}
