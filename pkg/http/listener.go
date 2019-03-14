package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var sslRequiredErrMsg = []byte("HTTP/1.1 403 Forbidden\r\n\r\nSSL required")

// HTTP methods.
var methods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
	"PRI", // HTTP 2
}

type acceptResult struct {
	conn net.Conn
	err error
}

type httpListener struct {
	mutex sync.Mutex
	tcpListeners []*net.TCPListener
	acceptCh chan acceptResult
	doneCh chan struct{}
	tlsConfig *tls.Config
	tcpKeepAliveTimeout time.Duration
	readTimeout time.Duration
	writeTimeout time.Duration
	maxHeaderBytes int
	updateBytesReadFunc func(int)
	updateBytesWrittenFunc func(int)
}

func isRoutineNetErr(err error) bool {
	if err == nil {
		return false
	}
	if nErr, ok := err.(*net.OpError); ok {
		if syscallErr, ok := nErr.Err.(*os.SyscallError); ok {
			if errno, ok := syscallErr.Err.(syscall.Errno); ok {
				return errno == syscall.ECONNRESET
			}
		}

		return nErr.Timeout()
	}
	return err == io.EOF || err.Error() == "EOF"
}

func (listener *httpListener) start() {
	listener.acceptCh = make(chan acceptResult)
	listener.doneCh   = make(chan struct{})

	send := func(result acceptCh, done <-chan struct{}) bool {
		select {
		case listener.acceptCh <- result:
			return true
		case <-doneCh:
			if result.conn != nil {
				result.conn.Close()
			}
			return false
		}
	}

	handleConn := func(tcpConn *net.TCPConn, doneCh <-chan struct{}) bool {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(listener.tcpKeepAliveTimeout)

		bufconn := newBufConn(tcpConn, listener.readTimeout, listener.writeTimeout, listener.maxHeaderBytes,
			listener.updateBytesReadFunc, listener.updateBytesWrittenFunc)
		if listener.tlsConfig != nil {
			ok, err := getPlainText(bufconn)
			if err != nil {
				if !isRoutineNetErr(err) {
					log.WithFields(log.Fields{
						"remoteAddr": bufconn.RemoteAddr().String(),
						"localAddr": bufconn.LocalAddr().String(),
					}).Error(err.Error())
				}
				bufconn.Close()
				return
			}

			if ok {
				bufconn.Write(sslRequiredErrMsg)
				bufconn.Close()
				return
			}
		}

		send(acceptResult{bufconn, nil}, doneCh)
	}

	handleListener := func(tcpListener *net.TCPListener, doneCh <-chan struct{}) {
		for {
			tcpConn, err := tcpListener.AcceptTCP()
			if err != nil {
				if !send(acceptResult{nil, err}, doneCh) {
					return
				}
			} else {
				go handleConn(tcpConn, doneCh)
			}
		}
	}

	for _, tcpListener := range listener.tcpListeners {
		go handleListener(tcpListener, listener.doneCh)
	}
}

func (listener *httpListener) Accept() (conn net.Conn, err error) {
	result, ok := <-listener.acceptCh
	if ok {
		return result.conn, result.err
	}

	return nil, syscall.EINVAL
}

func (listener *httpListener) Close() (err error) {
	listener.mutex.Lock()
	defer listener.mutex.Unlock()
	if listener.doneCh == nil {
		return syscall.EINVAL
	}

	for i := range listener.tcpListeners {
		listener.tcpListeners[i].Close()
	}
	close(listener.doneCh)

	listener.doneCh = nil
	return nil
}

func (listener *httpListener) Addr() (addr net.Addr) {
	addr = listener.tcpListeners[0].Addr()
	if len(listener.tcpListeners) == 1 {
		return addr
	}

	tcpAddr := addr.(*net.TCPAddr)
	if ip := net.ParseIP("0.0.0.0"); ip != nil {
		tcpAddr.IP = ip
	}

	addr = tcpAddr
	return addr
}

func (listener *httpListener) Addrs() (addr []net.Addr) {
	for i := range listener.tcpListeners {
		addrs = append(addrs, listener.tcpListeners[i].Addr())
	}

	return addrs
}

func newHTTPListener(serveAddrs []string,
	tlsConfig *tls.Config,
	tcpKeepAliveTimeout time.Duration,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	maxHeaderBytes int,
	updateBytesReadFunc func(int),
	updateBytesWrittenFunc func(int)) (listener *httpListener, err error) {

	var tcpListeners []*net.TCPListener

	defer func() {
		if err == nil {
			return
		}

		for _, tcpListener := range tcpListeners {
			tcpListener.Close()
		}
	}()

	for _, serverAddr := range serverAddrs {
		var l net.Listener
		if l, err = listen("tcp", serverAddr); err != nil {
			if l, err = fallbackListen("tcp", serverAddr); err != nil {
				return nil, err
			}
		}

		tcpListener, ok := l.(*net.TCPListener)
		if !ok {
			return nil, fmt.Errorf("unexpected listener type found %v, expected net.TCPListener", l)
		}

		tcpListeners = append(tcpListeners, tcpListener)
	}

	listener = &htpListener{
		tcpListeners: tcpListeners,
		tlsConfig: tlsConfig,
		tcpKeepAliveTimeout: tcpKeepAliveTimeout,
		readTimeout: readTimeout,
		writeTimeout: writeTimeout,
		maxHeaderBytes: maxHeaderBytes,
		updateBytesReadFunc: updateBytesReadFunc,
		updateBytesWrittenFunc: updateBytesWrittenFunc,
	}
	listener.start()

	return listener, nil
}
