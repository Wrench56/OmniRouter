package router

import (
	"context"
	"net"
	"omnirouter/internal/capabilities"
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

type HandlerTable struct {
	Handlers [methodCount]HTTPHandler
}

func execForMethodBit(fn func(int), method_mask uint8) {
	if method_mask == METHOD_UNKNOWN {
		return
	}

	if method_mask == METHOD_ANY {
		for i := range methodCount {
			fn(i)
		}
		return
	}

	for i := range methodCount {
		if method_mask&(1<<i) != 0 {
			fn(i)
		}
	}
}

type HTTPRouter interface {
	Register(caps capabilities.Capabilities, methodMask uint8, path string, h HTTPHandler) uint64
	Unregister(caps capabilities.Capabilities, methodMask uint8, path string) uint64
	Lookup(path string) (HandlerTable, bool)
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

	for len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}

	return p
}

func GetHTTPRouter() HTTPRouter {
	routerOnce.Do(func() {
		routerInst = NewHTTPRouter()
	})
	return routerInst
}

type routeEntry struct {
	table    HandlerTable
	wildcard bool
}

type radixRouter struct {
	mu   sync.RWMutex
	tree *radix.Tree
}

func NewHTTPRouter() HTTPRouter {
	return &radixRouter{tree: radix.New()}
}

func (r *radixRouter) getOrCreate(path string) *routeEntry {
	key := normalize(path)
	r.mu.RLock()
	if v, ok := r.tree.Get(key); ok {
		r.mu.RUnlock()
		return v.(*routeEntry)
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if v, ok := r.tree.Get(key); ok {
		return v.(*routeEntry)
	}
	re := &routeEntry{}
	r.tree.Insert(key, re)
	return re
}

func (r *radixRouter) Lookup(path string) (HandlerTable, bool) {
	raw := normalize(path)
	r.mu.RLock()
	if v, ok := r.tree.Get(raw); ok {
		out := v.(*routeEntry).table
		r.mu.RUnlock()
		return out, true
	}

	key, v, ok := r.tree.LongestPrefix(raw)
	r.mu.RUnlock()
	if !ok {
		return HandlerTable{}, false
	}
	re := v.(*routeEntry)
	if !re.wildcard {
		return HandlerTable{}, false
	}
	if key == "/" || raw == key || strings.HasPrefix(raw, key+"/") {
		return re.table, true
	}
	return HandlerTable{}, false
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
	go func() { serverErr <- s.Serve(ln) }()

	select {
	case <-ctx.Done():
		sdCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := s.ShutdownWithContext(sdCtx); err != nil {
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
	case "/favicon.ico", "/robots.txt":
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	path := string(ctx.Path())
	logger.Debug("Looking up handlers for path", "path", path)

	table, ok := GetHTTPRouter().Lookup(path)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	var methodBit uint8
	switch string(ctx.Method()) {
	case fasthttp.MethodGet:
		methodBit = METHOD_GET
	case fasthttp.MethodHead:
		methodBit = METHOD_HEAD
	case fasthttp.MethodPost:
		methodBit = METHOD_POST
	case fasthttp.MethodPut:
		methodBit = METHOD_PUT
	case fasthttp.MethodDelete:
		methodBit = METHOD_DELETE
	case fasthttp.MethodPatch:
		methodBit = METHOD_PATCH
	case fasthttp.MethodOptions:
		methodBit = METHOD_OPTIONS
	default:
		methodBit = METHOD_UNKNOWN
	}

	execForMethodBit(func(i int) {
		if table.Handlers[i] == nil {
			return
		}

		table.Handlers[i].Invoke(ContextPtr(ctx), RequestPtr(nil))
	}, methodBit)
}
