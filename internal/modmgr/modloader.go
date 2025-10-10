package modmgr

/*
#cgo CFLAGS: -I${SRCDIR}/bridges
#include "bridges/cffi.c"

#include <stdlib.h>
*/
import "C"

import (
	"omnirouter/internal/logger"
	"unsafe"
)

func loadModule(path string) {
	logger.Info("Checking modloaders...")
	if !C.health() {
		logger.Warn("C FFI module loader health check failed!")
	}
	logger.Info("Health checks done!")

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	C.load_module(cpath)
}
