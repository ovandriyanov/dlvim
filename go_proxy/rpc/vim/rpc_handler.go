package vim

import (
	"fmt"
	"log"
	"reflect"

	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
	_ "github.com/ovandriyanov/dlvim/go_proxy/upstream/command"
	"golang.org/x/xerrors"
)

const ServiceName = "Dlvim"

type RPCHandler struct {
	server *Server
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

func NewRPCHandler(server *Server) *RPCHandler {
	return &RPCHandler{server: server}
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

	command, err := upstream.NewCommand(stringArgv)
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
