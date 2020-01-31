package main

import (
	"time"

	"github.com/smallnest/ringbuffer"
	"github.com/smallnest/rpcx/protocol"
)

type connection struct {
	key           string
	buf           *ringbuffer.RingBuffer
	closeCallback func(err error)
	parseCallBack func(key string, msg *protocol.Message)

	found bool
	done  chan struct{}
}

func (c *connection) Start() {
	for {
		if !c.findFirstMsg() {
			continue
		}
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

const magicNumber byte = 0x08

func (c *connection) findFirstMsg() bool {
	if c.found {
		return true
	}

	buf := c.buf.Bytes()
	if len(buf) == 0 {
		time.Sleep(time.Millisecond)
		return false
	}
	if buf[0] == magicNumber {
		c.found = true
		return true
	}

	var index = len(buf) - 1

	for i, c := range buf {
		if c == magicNumber {
			index = i - 1
			break
		}
	}

	data := make([]byte, index)
	c.buf.Read(data)
	return false
}

func (c *connection) Close() {
	close(c.done)
}
