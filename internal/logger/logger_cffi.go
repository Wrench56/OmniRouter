//go:build cgo

package logger

/*
#include <stdint.h>
*/
import "C"

//export loginfo
func loginfo(msg *C.char, module *C.char) {
	SetLogCallerModule(C.GoString(module))
	Info(C.GoString(msg))
}

//export logwarn
func logwarn(msg *C.char, module *C.char) {
	SetLogCallerModule(C.GoString(module))
	Warn(C.GoString(msg))
}

//export logerror
func logerror(msg *C.char, module *C.char) {
	SetLogCallerModule(C.GoString(module))
	Error(C.GoString(msg))
}

//export logfatal
func logfatal(msg *C.char, module *C.char) {
	SetLogCallerModule(C.GoString(module))
	Fatal(C.GoString(msg))
}
