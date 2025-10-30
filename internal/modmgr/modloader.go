//go:build cgo

package modmgr

/*
#cgo CFLAGS: -I${SRCDIR}/bridges
#cgo windows LDFLAGS: -lkernel32
#include "bridges/cffi.h"
#include <stdlib.h>
*/
import "C"

import (
	"omnirouter/internal/logger"
	"omnirouter/internal/router"
	"unsafe"
)

type cHandler struct {
	fn    C.or_http_handler_t
	extra unsafe.Pointer
}

var _ router.HTTPHandler = cHandler{}

func (h cHandler) Invoke(ctx router.ContextPtr, req router.RequestPtr) {
	C.call_or_http_handler(
		h.fn,
		(*C.or_ctx_t)(ctx),
		(*C.or_http_req_t)(req),
		h.extra,
	)
}

func (mod *Module) Load() bool {
	cpath := C.CString(mod.path)
	defer C.free(unsafe.Pointer(cpath))
	mod.handle = C.cffi_load_module(cpath)
	return true
}

func (mod Module) Unload() bool {
	C.cffi_unload_module(mod.handle)
	return true
}

//export or_register_http
func or_register_http(path *C.char, handler C.or_http_handler_t, extra unsafe.Pointer) C.uint64_t {
	goPath := C.GoString(path)
	logger.Info("Added new HTTP handler for path ", "path", goPath)
	router.GetHTTPRouter().Register(goPath, cHandler{fn: handler, extra: extra})
	return C.uint64_t(1)
}

//export or_unregister_http
func or_unregister_http(path *C.char) {
	goPath := C.GoString(path)
	logger.Info("Unregistered HTTP handler for path", "path", goPath)
	router.GetHTTPRouter().Unregister(goPath)
}
