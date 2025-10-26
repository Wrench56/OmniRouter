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

func loadModule(path string) {
	logger.Info("Checking modloaders...")
	if !C.cffi_health() {
		logger.Warn("C FFI module loader health check failed!")
	}
	logger.Info("Health checks done!")

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	C.cffi_load_module(cpath)
}

//export or_register_http
func or_register_http(path *C.char, handler C.or_http_handler_t, extra unsafe.Pointer) C.uint64_t {
	goPath := C.GoString(path)
	logger.Info("Added new HTTP handler for path ", "path", goPath)
	router.GetHTTPRouter().Register(goPath, cHandler{fn: handler, extra: extra})
	return C.uint64_t(1)
}

