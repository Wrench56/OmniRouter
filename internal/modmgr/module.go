//go:build cgo

package modmgr

/*
#cgo CFLAGS: -I${SRCDIR}/bridges
#include "bridges/cffi.h"
*/
import "C"

import (
	"omnirouter/internal/capabilities"
	"path/filepath"
)

type Modtype int

const (
	MODTYPE_UNKNOWN Modtype = 0
	MODTYPE_DYNLIB  Modtype = 1
)

type ModuleI interface {
	Load() bool
	Unload() bool
	Stage() error
	Unstage() error
}

type Module struct {
	handle       C.mod_handle_t
	capabilities capabilities.Capabilities
	muid         MUID
	type_        Modtype
	path         string
	origPath     string
	filename     string
}

func extensionToModuleType(ext string) Modtype {
	switch ext {
	case ".so", ".dll", ".dylib":
		return MODTYPE_DYNLIB
	default:
		return MODTYPE_UNKNOWN
	}
}

func IsModuleFile(filename string) bool {
	return extensionToModuleType(filepath.Ext(filename)) != MODTYPE_UNKNOWN
}
