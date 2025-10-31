//go:build cgo

package modmgr

/*
#cgo CFLAGS: -I${SRCDIR}/bridges
#include "bridges/cffi.h"
#include <stdlib.h>
*/
import "C"

import (
	"omnirouter/internal/logger"
	"omnirouter/internal/router"
	"unsafe"
)

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
