package modmgr

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"omnirouter/internal/logger"

	"github.com/sashka/atomicfile"
)

var (
	mirrorMu  sync.Mutex
	src2mod   = make(map[string]Module)
	mirrordir string
)

func SetMirrorDir(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return err
	}
	mirrorMu.Lock()
	mirrordir = abs
	mirrorMu.Unlock()
	return nil
}

func CreateModule(path string) {
	if !IsModuleFile(filepath.Base(path)) {
		return
	}

	mod := Module{
		handle:   0,
		type_:    extensionToModuleType(filepath.Ext(filepath.Base(path))),
		path:     path,
		filename: filepath.Base(path),
	}

	mirrorMu.Lock()
	src2mod[filepath.Clean(path)] = mod
	mirrorMu.Unlock()
	mod.Stage()
}

func ReloadModule(path string) {
	if !IsModuleFile(filepath.Base(path)) {
		return
	}

	mirrorMu.Lock()
	mod, ok := src2mod[filepath.Clean(path)]
	if ok {
		err := mod.Unstage()
		if err != nil {
			logger.Error("Unable to unload module with", "path", path)
		}

		dst := filepath.Join(mirrordir, mod.filename)
		mode := os.FileMode(0o644)
		if fi, err := os.Stat(mod.path); err == nil {
			mode = fi.Mode()
		}

		if err := copyFileAtomic(mod.path, dst, mode); err != nil {
			logger.Error("Could not copy file atomically", "src", mod.path, "dst", dst)
		}

		mod.Load()
		logger.Info("Staged module", "path", mod.path, "type", mod.type_)

		mirrorMu.Unlock()

	} else {
		mirrorMu.Unlock()
		CreateModule(path)
	}

}

func RemoveModule(path string) {
	if !IsModuleFile(filepath.Base(path)) {
		return
	}

	mirrorMu.Lock()
	mod, ok := src2mod[filepath.Clean(path)]
	if ok {
		err := mod.Unstage()
		if err == nil {
			delete(src2mod, path)
		} else {
			logger.Error("Unable to unload module with", "path", path)
		}
	} else {
		logger.Error("Unalbe to find and unload module with", "path", path)
	}
	mirrorMu.Unlock()
}

func (mod *Module) Stage() error {
	dst := filepath.Join(mirrordir, mod.filename)

	mode := os.FileMode(0o644)
	if fi, err := os.Stat(mod.path); err == nil {
		mode = fi.Mode()
	}

	if err := copyFileAtomic(mod.path, dst, mode); err != nil {
		return err
	}

	mod.Load()
	logger.Info("Staged module", "path", mod.path, "type", mod.type_)
	return nil
}

func (mod *Module) Unstage() error {
	mod.Unload()
	err := removeFileAtomic(filepath.Join(mirrordir, mod.filename))
	if err != nil {
		return err
	}
	logger.Info("Unstaged module", "path", mod.path, "type", mod.type_)
	return nil
}

func copyFileAtomic(src string, dst string, mode os.FileMode) error {
	sf, err := os.Open(src)
	if err != nil {
		logger.Error("Failed to open source", "src", src, "err", err)
		return err
	}
	defer sf.Close()

	f, err := atomicfile.New(dst, mode)
	if err != nil {
		logger.Error("Failed to open atomic file", "dst", dst, "err", err)
		return err
	}
	defer f.Abort()

	if _, err := io.Copy(f, sf); err != nil {
		logger.Error("Copy failed", "src", src, "dst", dst, "err", err)
		return err
	}
	if err := f.Close(); err != nil {
		logger.Error("Commit failed", "dst", dst, "err", err)
		return err
	}

	logger.Info("Copied shared library atomically", "src", src, "dst", dst)
	return nil
}

/* Technically not atomic on all FS, but whatever, I'm keeping the name */
func removeFileAtomic(src string) error {
	if err := os.Remove(src); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		logger.Error("Failed to remove file", "path", src, "err", err)
		return err
	}
	return nil
}
