package router

import (
	"strings"
	"sync"
	"unsafe"

	radix "github.com/armon/go-radix"
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

func Setup() {
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
