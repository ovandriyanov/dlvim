package main

import (
	"context"
	"os"
	"sync"
)

type ioResult struct {
	n   int
	err error
}

type StdioConn struct {
	readQueue  chan []byte
	readDone   chan ioResult
	writeQueue chan []byte
	writeDone  chan ioResult
	closed     chan struct{}
	wg         sync.WaitGroup
}

func (c *StdioConn) Read(p []byte) (n int, err error) {
	select {
	case <-c.closed:
		return 0, context.Canceled
	case c.readQueue <- p:
	}

	select {
	case res := <-c.readDone:
		return res.n, res.err
	case <-c.closed:
		return 0, context.Canceled
	}
}

func (c *StdioConn) Write(p []byte) (n int, err error) {
	select {
	case <-c.closed:
		return 0, context.Canceled
	case c.writeQueue <- p:
	}

	select {
	case res := <-c.writeDone:
		return res.n, res.err
	case <-c.closed:
		return 0, context.Canceled
	}
}

func (c *StdioConn) Close() error {
	select {
	case <-c.closed:
	default:
		close(c.closed)
		//c.wg.Wait()
	}
	return nil
}

func (c *StdioConn) doIO(requests chan []byte, results chan ioResult, ioFunc func([]byte) (int, error)) {
	for {
		select {
		case req := <-requests:
			n, err := ioFunc(req)
			select {
			case results <- ioResult{n: n, err: err}:
			case <-c.closed:
				return
			}
		case <-c.closed:
			return
		}
	}
}

func NewStdioConn() *StdioConn {
	conn := &StdioConn{
		readQueue:  make(chan []byte),
		readDone:   make(chan ioResult),
		writeQueue: make(chan []byte),
		writeDone:  make(chan ioResult),
		closed:     make(chan struct{}),
		wg:         sync.WaitGroup{},
	}
	conn.wg.Add(2)
	go func() {
		defer conn.wg.Done()
		conn.doIO(conn.readQueue, conn.readDone, os.Stdin.Read)
	}()
	go func() {
		defer conn.wg.Done()
		conn.doIO(conn.writeQueue, conn.writeDone, os.Stdout.Write)
	}()
	return conn
}
