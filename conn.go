package main

import (
	"github.com/smallnest/ringbuffer"
	"github.com/smallnest/rpcx/protocol"
)

type connection struct {
	key           string
	buf           *ringbuffer.RingBuffer
	closeCallback func(err error)
	parseCallBack func(key string, msg *protocol.Message)

	done chan struct{}
}

func (c *connection) Start() {
	for {
		select {
		case <-c.done:
			return
		default:
			msg, err := protocol.Read(c.buf)
			if err != nil {
				c.closeCallback(err)
				return
			}
			c.parseCallBack(c.key, msg)
		}

	}
}

func (c *connection) Close() {
	close(c.done)
}
