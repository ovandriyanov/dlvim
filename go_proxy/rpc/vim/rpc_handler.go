package vim

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
	_ "github.com/ovandriyanov/dlvim/go_proxy/upstream/command"
	"github.com/ovandriyanov/dlvim/go_proxy/vimevent"
	"golang.org/x/xerrors"
)

const ServiceName = "Dlvim"

type RPCHandler struct {
	server *Server
	ctx    context.Context
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

func (h *RPCHandler) GetNextEvent(req map[string]interface{}, resp *map[string]interface{}) error {
	var event vimevent.Event
	select {
	case event = <-h.server.events:
	case <-h.ctx.Done():
		return context.Canceled
	}

	(*resp)["kind"] = event.Kind()
	(*resp)["payload"] = event
	return nil
}

func NewRPCHandler(server *Server, ctx context.Context) *RPCHandler {
	return &RPCHandler{
		server: server,
		ctx:    ctx,
	}
}
