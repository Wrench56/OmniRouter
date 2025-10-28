package modmgr

import (
	"path/filepath"
)

type Modtype int
const (
	MODTYPE_UNKNOWN Modtype = 0
	MODTYPE_DYNLIB Modtype = 1
)

type ModuleI interface {
	Load() bool
	Unload() bool
	Stage() error
	Unstage() error
}

type Module struct {
	handle uintptr
	type_ Modtype
	path string
	filename string
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
