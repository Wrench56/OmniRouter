package modmgr

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sashka/atomicfile"
	"omnirouter/internal/logger"
)

var isSoLikeRegex = regexp.MustCompile(`(?i)\.(dll|dylib|so(\.[0-9]+)*)$`)

func isSoLike(name string) bool {
	return isSoLikeRegex.MatchString(filepath.Base(name))
}

func CopyMods(src string, dst string) error {
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		logger.Error("Failed to resolve source path", "path", src, "err", err)
		return err
	}

	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		logger.Error("Failed to resolve destination path", "path", dst, "err", err)
		return err
	}

	if err := os.MkdirAll(dstAbs, 0o755); err != nil {
		logger.Error("Failed to create destination root", "dst", dstAbs, "err", err)
		return err
	}

	sep := string(os.PathSeparator)
	dstInsideSrc := strings.HasPrefix(dstAbs+sep, srcAbs+sep)

	logger.Info("Copying shared libs", "src", srcAbs, "dst", dstAbs)

	return filepath.WalkDir(srcAbs, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if d != nil && d.IsDir() {
				logger.Warn("Skipping unreadable directory", "path", p, "err", walkErr)
				return fs.SkipDir
			}
			logger.Warn("Skipping entry due to walk error", "path", p, "err", walkErr)
			return nil
		}

		if d == nil {
			return nil
		}

		if d.IsDir() && dstInsideSrc && p == dstAbs {
			return fs.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		if !d.Type().IsRegular() || !isSoLike(p) {
			return nil
		}

		dstPath := filepath.Join(dstAbs, filepath.Base(p))

		fi, err := d.Info()
		var mode os.FileMode = 0o644
		if err == nil {
			mode = fi.Mode()
		} else {
			logger.Debug("Stat failed; using default mode", "path", p, "err", err)
		}

		err = copyFileAtomic(p, dstPath, mode)
		if err != nil {
			logger.Warn("Failed to copy shared lib", "src", p, "dst", dstPath, "err", err)
			return nil
		}
		return nil
	})
}

func copyFileAtomic(src, dst string, mode os.FileMode) error {
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
