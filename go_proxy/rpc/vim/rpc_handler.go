package vim

import (
	"fmt"
	"log"
	"reflect"
)

const ServiceName = "Dlvim"

type RPCHandler struct{}

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

func NewRPCHandler() *RPCHandler {
	return &RPCHandler{}
}

func (h *RPCHandler) Foo(req map[string]interface{}, resp *map[string]interface{}) error {
	(*resp)["foo"] = "bar"
	return nil
}

func (h *RPCHandler) Initialize(req map[string]interface{}, resp *map[string]interface{}) error {
	return nil
}
