//go:build cgo

package logger

/*
#include <stdint.h>
*/
import "C"

//export or_loginfo
func or_loginfo(msg *C.char, module *C.char) {
	SetLogCallerModule(C.GoString(module))
	Info(C.GoString(msg))
}

//export or_logwarn
func or_logwarn(msg *C.char, module *C.char) {
	SetLogCallerModule(C.GoString(module))
	Warn(C.GoString(msg))
}

//export or_logerror
func or_logerror(msg *C.char, module *C.char) {
	SetLogCallerModule(C.GoString(module))
	Error(C.GoString(msg))
}

//export or_logfatal
func or_logfatal(msg *C.char, module *C.char) {
	SetLogCallerModule(C.GoString(module))
	Fatal(C.GoString(msg))
}
