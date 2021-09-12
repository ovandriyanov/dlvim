package vim

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/rpc"

	"golang.org/x/xerrors"
)

type VimRPCCodec struct {
	vimConn        io.ReadWriteCloser
	decoder        *json.Decoder
	encoder        *json.Encoder
	methodName     string
	requestMessage json.RawMessage
}

func (v *VimRPCCodec) Close() error {
	return v.vimConn.Close()
}

var k json.Unmarshaler

func (v *VimRPCCodec) ReadRequestHeader(request *rpc.Request) (err error) {
	defer func() {
		if err != nil {
			log.Printf("Cannot read request header: %s\n", err.Error())
		}
	}()

	var methodName *string
	var requestMessage *json.RawMessage
	req := [2]interface{}{methodName, requestMessage}

	var reqID *uint64
	message := [2]interface{}{reqID, &req}

	if err = v.decoder.Decode(&message); err != nil {
		err = xerrors.Errorf("Request decoding failed: %w", err)
		return
	}

	if reqID == nil {
		err = xerrors.New("Request ID not provided")
	}

	if methodName == nil {
		err = xerrors.New("Method name not provided")
	}
	v.methodName = *methodName

	if requestMessage == nil {
		v.requestMessage = []byte("null")
	} else {
		v.requestMessage = *requestMessage
	}

	log.Printf("Dlvim method called: %s\n", v.methodName)
	request.ServiceMethod = fmt.Sprintf("Dlvim.%s", v.methodName)
	request.Seq = *reqID
	return
}

func (v *VimRPCCodec) ReadRequestBody(body interface{}) error {
	if body == nil {
		return nil
	}
	if err := json.Unmarshal(v.requestMessage, body); err != nil {
		return xerrors.Errorf("%s: cannot unmarshal request message into %T: %w", v.methodName, body, err)
	}
	return nil
}

func (v *VimRPCCodec) WriteResponse(response *rpc.Response, body interface{}) error {
	var responseBody interface{}
	if response.Error != "" {
		responseBody = map[string]string{"Error": response.Error}
	} else {
		responseBody = body
	}

	responseMessage := [2]interface{}{response.Seq, responseBody}
	if err := v.encoder.Encode(responseMessage); err != nil {
		return xerrors.Errorf("%s: cannot encode response message of type %T: %w", response.ServiceMethod, body, err)
	}
	return nil
}

func NewVimRPCCodec(vimConn io.ReadWriteCloser) VimRPCCodec {
	decoder := json.NewDecoder(vimConn)
	decoder.UseNumber()
	encoder := json.NewEncoder(vimConn)
	return VimRPCCodec{
		vimConn: vimConn,
		decoder: decoder,
		encoder: encoder,
	}
}
