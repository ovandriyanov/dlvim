package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/rpc"

	"golang.org/x/xerrors"
)

type req [2]json.RawMessage

type VimRPCCodec struct {
	vimConn io.ReadWriteCloser
	decoder *json.Decoder
	req
}

func (v *VimRPCCodec) Close() error {
	return v.vimConn.Close()
}

func (v *VimRPCCodec) ReadRequestHeader(request *rpc.Request) (err error) {
	defer func() {
		if err != nil {
			log.Printf("Cannot read request header: %s\n", err.Error())
		}
	}()

	var reqID *uint64
	var req []interface{}
	message := [2]interface{}{reqID, &req}
	if err = v.decoder.Decode(&message); err != nil {
		err = xerrors.Errorf("Request decoding failed: %w", err)
		return
	}

	if reqID == nil {
		err = xerrors.New("Request ID not provided")
	}

	if len(req) < 1 {
		err = xerrors.New("Method name not provided")
		return
	}

	methodName, ok := req[0].(string)
	if !ok {
		err = xerrors.Errorf("Method name is not a string but %T", req[0])
		return
	}

	log.Printf("Dlvim method called: %s\n", methodName)
	request.ServiceMethod = fmt.Sprintf("Dlvim.%s", methodName)
	request.Seq = *reqID
	return
}

//func (v *VimRPCCodec) ReadRequestBody(body interface{}) error {
//
//}
//
//func (v *VimRPCCodec) WriteResponse(response *rpc.Response, body interface{}) error {
//
//}

func NewVimRPCCodec(vimConn io.ReadWriteCloser) VimRPCCodec {
	decoder := json.NewDecoder(vimConn)
	decoder.UseNumber()
	return VimRPCCodec{
		vimConn: vimConn,
		decoder: decoder,
	}
}
