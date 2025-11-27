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

	muid := generateMUID64(mod)
	mod.muid = muid
	mod.handle = C.cffi_load_module(cpath, C.muid_t(mod.muid))
	return true
}

func (mod Module) Unload() bool {
	C.cffi_unload_module(mod.handle, C.muid_t(mod.muid))
	return true
}
