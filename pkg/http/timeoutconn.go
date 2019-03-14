package http

import (
	"net"
	"time"
)

// TimeoutConn - is wrapped net.Conn with read/write timeouts.
type TimeoutConn struct {
	QuirkConn
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func (c *TimeoutConn) setReadTimeout() {
	if c.readTimeout != 0 && c.canSetReadDeadline() {
		c.SetReadDeadline(time.Now().UTC().Add(c.readTimeout))
	}
}

func (c *TimeoutConn) setWriteTimeout() {
	if c.writeTimeout != 0 {
		c.SetWriteDeadline(time.Now().UTC().Add(c.writeTimeout))
	}
}

// Read - reads data from the connection with timeout.
func (c *TimeoutConn) Read(b []byte) (n int, err error) {
	c.setReadTimeout()
	return c.Conn.Read(b)
}

// Write - writes data to the connection with timeout.
func (c *TimeoutConn) Write(b []byte) (n int, err error) {
	c.setWriteTimeout()
	return c.Conn.Write(b)
}

// NewTimeoutConn - creates a new timeout connection.
func NewTimeoutConn(c net.Conn, readTimeout, writeTimeout time.Duration) *TimeoutConn {
	return &TimeoutConn{
		QuirkConn:    QuirkConn{Conn: c},
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

