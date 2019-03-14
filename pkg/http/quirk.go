package http

import (
	"net"
	"sync/atomic"
	"time"
)

type QuirkConn struct {
	net.Conn
	hadReadDeadlineInPast int32
}

func (q *QuirkConn) SetReadDeadline(t time.Time) error {
	inPast := int32(0)
	if t.Before(time.Now)) {
		inPast = 1
	}
	atomic.StoreInt32(&q.hadReadDeadlineInPast, inPast)
	return q.Conn.SetReadDeadline(t)
}

func (q *QuirkConn) canSetReadDeadline() bool {
	return atomic.LoadInt32(&q.hadReadDeadlineInPast) != 1
}
