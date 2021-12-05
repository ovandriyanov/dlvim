package vim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/rpc"

	"golang.org/x/xerrors"
)

type RPCCodec struct {
	vimConn        io.ReadWriteCloser
	decoder        *json.Decoder
	encoder        *json.Encoder
	requestMessage requestMessage
}

func (v *RPCCodec) Close() error {
	return v.vimConn.Close()
}

type requestMessage struct {
	seq     uint64
	method  string
	payload json.RawMessage
}

func (m *requestMessage) delim(decoder *json.Decoder, delim json.Delim) error {
	tok, err := decoder.Token()
	if err != nil {
		return err
	}
	if tok != delim {
		return xerrors.Errorf("Expected %c, got %v", delim, tok)
	}
	return nil
}

func (m *requestMessage) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := m.delim(decoder, '['); err != nil {
		return xerrors.Errorf("Request array: %w", err)
	}
	if err := decoder.Decode(&m.seq); err != nil {
		return xerrors.Errorf("Request sequence number: %w", err)
	}
	if err := m.delim(decoder, '['); err != nil {
		return xerrors.Errorf("Payload array: %w", err)
	}
	if err := decoder.Decode(&m.method); err != nil {
		return xerrors.Errorf("Method name: %w", err)
	}
	if err := decoder.Decode(&m.payload); err != nil {
		return xerrors.Errorf("Payload: %w", err)
	}
	if err := m.delim(decoder, ']'); err != nil {
		return xerrors.Errorf("Payload array: %w", err)
	}
	if err := m.delim(decoder, ']'); err != nil {
		return xerrors.Errorf("Request array: %w", err)
	}
	return nil
}

func (v *RPCCodec) ReadRequestHeader(request *rpc.Request) (err error) {
	defer func() {
		if err != nil {
			log.Printf("Cannot read request header: %s\n", err.Error())
		}
	}()

	if err = v.decoder.Decode(&v.requestMessage); err != nil {
		err = xerrors.Errorf("Cannot decode request message: %w", err)
		return
	}

	request.ServiceMethod = fmt.Sprintf("Dlvim.%s", v.requestMessage.method)
	request.Seq = v.requestMessage.seq
	return
}

func (v *RPCCodec) ReadRequestBody(body interface{}) (err error) {
	defer func() {
		if err != nil {
			log.Printf("Cannot read request body for method %s (seq %d): %s\n", v.requestMessage.method, v.requestMessage.seq, err.Error())
			return
		}
		log.Printf("Vim: call %s with argument %v\n", v.requestMessage.method, body)
	}()
	if body == nil {
		return
	}
	if err = json.Unmarshal(v.requestMessage.payload, body); err != nil {
		err = xerrors.Errorf("%s (seq %d): cannot unmarshal request message into %T: %w", v.requestMessage.method, v.requestMessage.seq, body, err)
		return
	}
	return
}

func (v *RPCCodec) WriteResponse(response *rpc.Response, body interface{}) error {
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

func NewRPCCodec(vimConn io.ReadWriteCloser) *RPCCodec {
	decoder := json.NewDecoder(vimConn)
	decoder.UseNumber()
	encoder := json.NewEncoder(vimConn)
	return &RPCCodec{
		vimConn: vimConn,
		decoder: decoder,
		encoder: encoder,
	}
}
