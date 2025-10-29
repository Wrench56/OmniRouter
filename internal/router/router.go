package router

import (
	"context"
	"net"
	"omnirouter/internal/logger"
	"strings"
	"sync"
	"time"
	"unsafe"

	radix "github.com/armon/go-radix"
	"github.com/valyala/fasthttp"
)

var (
	routerOnce sync.Once
	routerInst HTTPRouter
)

type ContextPtr unsafe.Pointer
type RequestPtr unsafe.Pointer

type HTTPHandler interface {
	Invoke(ctx ContextPtr, req RequestPtr)
}

type HTTPRouter interface {
	Register(path string, h HTTPHandler)
	Lookup(path string) (HTTPHandler, bool)
}

func setup() {
	routerOnce.Do(func() {
		routerInst = &radixRouter{tree: radix.New()}
	})
}

func normalize(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return p
}

func GetHTTPRouter() HTTPRouter {
	routerOnce.Do(func() {
		routerInst = &radixRouter{tree: radix.New()}
	})
	return routerInst
}

type radixRouter struct {
	mu   sync.RWMutex
	tree *radix.Tree
}

func NewHTTPRouter() HTTPRouter {
	return &radixRouter{tree: radix.New()}
}

func (r *radixRouter) Register(path string, h HTTPHandler) {
	p := normalize(path)

	if strings.HasSuffix(p, "*") {
		p = strings.TrimSuffix(p, "*")
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
	}

	r.mu.Lock()
	r.tree.Insert(p, h)
	r.mu.Unlock()
}

func (r *radixRouter) Lookup(path string) (HTTPHandler, bool) {
	p := normalize(path)

	r.mu.RLock()
	if v, ok := r.tree.Get(p); ok {
		r.mu.RUnlock()
		return v.(HTTPHandler), true
	}

	_, v, ok := r.tree.LongestPrefix(p)
	r.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return v.(HTTPHandler), true
}

func startServer(addr string) (*fasthttp.Server, net.Listener, error) {
	setup()

	s := &fasthttp.Server{
		Handler:                       dispatch,
		NoDefaultServerHeader:         true,
		NoDefaultDate:                 true,
		DisableHeaderNamesNormalizing: true,
		ReadTimeout:                   10 * time.Second,
		WriteTimeout:                  20 * time.Second,
		MaxRequestBodySize:            16 << 20,
		TCPKeepalive:                  true,
		ReadBufferSize:                4096,
		WriteBufferSize:               4096,
	}

	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		return nil, nil, err
	}
	return s, ln, nil
}

func RunServer(ctx context.Context, addr string) error {
	logger.Info("Running server on address", "addr", addr)
	s, ln, err := startServer(addr)
	if err != nil {
		return err
	}
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- s.Serve(ln)
	}()

	select {
	case <-ctx.Done():
		sdCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := s.ShutdownWithContext(sdCtx)
		if err != nil {
			logger.Warn("Server could not be shut down cleanly", "err", err)
		}

		if sdCtx.Err() == context.DeadlineExceeded {
			_ = ln.Close()
		}

		select {
		case <-serverErr:
		case <-time.After(1 * time.Second):
			logger.Warn("Server did not stop gracefully in time")
		}
		return ctx.Err()
	case err := <-serverErr:
		return err
	}
}

func dispatch(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/favicon.ico":
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	case "/robots.txt":
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	path := string(ctx.Path())
	logger.Debug("Looking up handler for path", "path", path)

	h, ok := GetHTTPRouter().Lookup(path)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	h.Invoke(ContextPtr(unsafe.Pointer(ctx)), RequestPtr(nil))
}
