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
	"crypto/rand"
	"encoding/binary"
	"omnirouter/internal/logger"
	"omnirouter/internal/router"
	"sync"
	"unsafe"
)

var (
	muidMap sync.Map
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

func InitMUID64Map() {
	muidMap.Store(0, 0)
}

func generateMUID64() uint64 {
	var b [8]byte
	for i := 0; i < 10; i++ {
		_, err := rand.Read(b[:])
		if err != nil {
			continue
		}
		muid := binary.BigEndian.Uint64(b[:])
		_, ok := muidMap.Load(muid)
		if !ok {
			muidMap.Store(muid, 0)
			return muid
		}
	}

	logger.Error("Couldn't create unique module ID in 10 tries, returning invalid UUID")
	return 0
}

func (mod *Module) Load() bool {
	cpath := C.CString(mod.path)
	defer C.free(unsafe.Pointer(cpath))
	
	muid := generateMUID64()
	mod.muid = muid
	mod.handle = C.cffi_load_module(cpath, C.muid_t(mod.muid))
	return true
}

func (mod Module) Unload() bool {
	C.cffi_unload_module(mod.handle)
	return true
}
