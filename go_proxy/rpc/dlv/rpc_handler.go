package dlv

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"reflect"
)

const ServiceName = "RPCServer" // Dlv client expects that service name

var KnownMethods map[string]struct{}

var debugRPC = flag.Bool("debug-rpc", false, "Show full requests and responses sent between dlv and the proxy")

func fqmn(method string) string {
	return fmt.Sprintf("%s.%s", ServiceName, method)
}

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

type RPCHandler struct {
	dlvClient *rpc.Client
}

func (h *RPCHandler) defaultHandler(method string, req map[string]interface{}, resp *map[string]interface{}) error {
	if *debugRPC {
		jsonReq, _ := json.MarshalIndent(req, "", "    ")
		log.Printf("Request: %s\n", jsonReq)
	}
	err := h.dlvClient.Call(method, req, resp)
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
	if *debugRPC {
		jsonResp, _ := json.MarshalIndent(resp, "", "    ")
		log.Printf("Response: %s\n", jsonResp)
	}
	return err
}

func NewRPCHandler(dlvClient *rpc.Client) *RPCHandler {
	return &RPCHandler{dlvClient: dlvClient}
}
