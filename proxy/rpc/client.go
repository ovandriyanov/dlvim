package rpc

import (
	"io"
)

type Client interface {
	io.Closer
	Call(serviceMethod string, args interface{}, reply interface{}) error
}
