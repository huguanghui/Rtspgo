package rtspgo

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// connReader 用于处理连接中的读
type connReader struct {
	conn    *conn
	mu      sync.Mutex
	hasByte bool
	byteBuf [1]byte
	cond    *sync.Cond
	inRead  bool
	aborted bool
	remain  int64
}

var (
	bufioReaderPool   sync.Pool
	bufioWriter2kPool sync.Pool
	bufioWriter4kPool sync.Pool
)

func bufioWriterPool(size int) *sync.Pool {
	switch size {
	case 2 << 10:
		return &bufioWriter2kPool
	case 4 << 10:
		return &bufioWriter4kPool
	}
	return nil
}

func newBufioReader(r io.Reader) *bufio.Reader {
	if v := bufioReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReader(r)
}

func newBufioWriterSize(w io.Writer, size int) *bufio.Writer {
	pool := bufioWriterPool(size)
	if pool != nil {
		if v := pool.Get(); v != nil {
			bw := v.(*bufio.Writer)
			bw.Reset(w)
			return bw
		}
	}
	return bufio.NewWriterSize(w, size)
}

// conn 显示rtsp中连接中的信息
type conn struct {
	server *Server

	cancelCtx context.CancelFunc

	rwc        net.Conn
	remoteAddr string
	werr       error

	r *connReader

	bufr *bufio.Reader
	bufw *bufio.Writer

	mu        sync.Mutex
	hijackedv bool
}

// func (c *conn) setState(nc net.Conn, state ConnState) {
// 	src := c.server
// 	switch state {
// 		case
// 	}
// }

// 为连接创建服务
func (c *conn) serve(ctx context.Context) {
	c.remoteAddr = c.rwc.RemoteAddr().String()
	ctx = context.WithValue(ctx, LocalAddrContextKey, c.rwc.LocalAddr)

	ctx, cancelCtx := context.WithCancel(ctx)
	c.cancelCtx = cancelCtx
	defer cancelCtx()

	c.r = &connReader{conn: c}
	c.bufr = newBufioReader(c.r)
	c.bufw = newBufioWriterSize(checkConnErrorWriter{c}, 4<<10)

	for {

	}
}

//
func (c *conn) readRequest(ctx context.Context) (w *response, err error) {
	
}

var (
	// ServerContextKey comment
	ServerContextKey = &contextKey{"http-server"}
	// LocalAddrContextKey comment
	LocalAddrContextKey = &contextKey{"local-addr"}
)

// Server RTSP服务的对象
type Server struct {
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	ErrorLog          *log.Logger

	inShutdown int32

	mu         sync.Mutex
	listeners  map[*net.Listener]struct{}
	doneChan   chan struct{}
	onShutdown []func()
}

// testHookServerServe 在Server开始时设置一个钩子
var testHookServerServe func(*Server, net.Listener)

// ErrServerClosed 当RTSP服务被关闭后的错误返回
var ErrServerClosed = errors.New("http: Server closed")

func (s *Server) getDoneChan() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getDoneChanLocked()
}

func (s *Server) getDoneChanLocked() chan struct{} {
	if s.doneChan == nil {
		s.doneChan = make(chan struct{})
	}
	return s.doneChan
}

func (s *Server) closeDoneChanLocked() {
	ch := s.getDoneChanLocked()
	select {
	case <-ch:
	default:
		close(ch)
	}
}

// ListenAndServe 接口是rtsp服务内部的接口
func (s *Server) ListenAndServe() error {
	if s.shuttingDown() {
		return ErrServerClosed
	}
	addr := s.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
}

// Serve 接受l监听的连接,每来个连接创建一个新的协程,调用s.Handler来处理
func (s *Server) Serve(l net.Listener) error {
	if fn := testHookServerServe; fn != nil {
		fn(s, l)
	}
	l = &onceCloseListener{Listener: l}
	defer l.Close()

	// if !s.trackListener(&l, true) {
	// 	return ErrServerClosed
	// }
	// defer s.trackListener(&l, false)
	var tempDelay time.Duration
	baseCtx := context.Background()
	ctx := context.WithValue(baseCtx, ServerContextKey, s)
	for {
		rw, e := l.Accept()
		if e != nil {
			select {
			case <-s.getDoneChan():
				return ErrServerClosed
			default:
			}
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		c := s.newConn(rw)
		//c.setState(c.rwc, StateNew)
		go c.serve(ctx)
	}
}

// LocalHandleConn 测试TCP处理
func LocalHandleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		fmt.Println("Start to read from conn ...")
		_, e := conn.Read(buf)
		if e != nil {
			continue
		}
		conn.Write([]byte("response"))
	}
}

// ListenAndServe 对外提供一个创建RTSP服务的接口
func ListenAndServe(addr string) error {
	server := &Server{Addr: addr}
	return server.ListenAndServe()
}

// shuttingDown 用于检测当前是服务是不是在关闭状态
func (s *Server) shuttingDown() bool {
	return atomic.LoadInt32(&s.inShutdown) != 0
}

// trackListener
func (s *Server) trackListener(ln *net.Listener, add bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listeners == nil {
		s.listeners = make(map[*net.Listener]struct{})
	}
	if add {
		if s.shuttingDown() {
			return false
		}
		s.listeners[ln] = struct{}{}
	} else {
		delete(s.listeners, ln)
	}
	return true
}

//const debugServerConnections = false

// newConn 创建新的连接
func (s *Server) newConn(rwc net.Conn) *conn {
	c := &conn{
		server: s,
		rwc:    rwc,
	}
	// if debugServerConnections {
	// 	c.rwc =
	// }
	return c
}

// tcpKeepAliveListener 用于设置TCP的keep-alive超时时间去接收连接
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (t tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := t.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

type onceCloseListener struct {
	net.Listener
	once     sync.Once
	closeErr error
}

func (o *onceCloseListener) Close() error {
	o.once.Do(o.close)
	return o.closeErr
}

func (o *onceCloseListener) close() {
	o.closeErr = o.Listener.Close()
}

type checkConnErrorWriter struct {
	c *conn
}

func (c checkConnErrorWriter) Write(p []byte) (n int, err error) {
	n, err = c.c.rwc.Write(p)
	if err != nil && c.c.werr == nil {
		c.c.werr = err
		c.c.cancelCtx()
	}
	return
}
