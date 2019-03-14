package http

import (
	"bufio"
	"net"
	"time"
)

type BufConn struct {
	QuirkConn
	bufReader *bufio.Reader
	readTimeout time.Duration
	writeTimeout time.Duration
	updateBytesReadFunc func(int)
	updateBytesWrittenFunc func(int)
}

func (c *BufConn) setReadTimeout() {
	if c.readTimeout != 0 && c.canSetReadDeadline() {
		c.SetReadDeadline(time.Now().UTC().Add(c.readTimeout))
	}
}

func (c *BufConn) setWriteTimeout() {
	if c.writeTimeout != 0 {
		c.SetWriteDeadline(time.Now().UTC().Add(c.writeTimeout))
	}
}

func (c *BufConn) RemoveTimeout() {
	c.readTimeout = 0
	c.writeTimeout = 0

	c.SetDeadline(time.Time{})
}

func (c *BufConn) Peek(n int) ([]byte, error) {
	c.setReadTimeout()
	return c.bufReader.Peek(n)
}

func (c *BufConn) Read(b []byte) (n int, err error) {
	c.setReadTimeout()
	n, err = c.bufReader.Read(b)
	if err == nil && c.updateBytesReadFunc != nil {
		c.updateBytesReadFunc(n)
	}

	return n, err
}

func (c *BufConn) Write(b []byte) (n int, err error) {
	c.setWriteTimeout()
	n, err = c.Conn.Write(b)
	if err == nil && c.updateBytesWrittenFunc != nil {
		c.updateBytesWrittenFunc(n)
	}

	return n, err
}

func newBufConn(c net.Conn, readTimeout, writeTimeout time.Duration, maxHeaderBytes int, updateBytesReadFunc, updateBytesWrittenFunc func(int)) *BufConn {
	return &BufConn{
		QuirkConn: QuirkConn{Conn: c},
		bufReader: bufio.NewReaderSize(c, maxHeaderBytes),
		readTimeout: readTimeout,
		writeTimeout: writeTimeout,
		updateBytesReadFunc: updateBytesReadFunc,
		updateBytesWrittenFunc: updateBytesWrittenFunc,
	}
}
