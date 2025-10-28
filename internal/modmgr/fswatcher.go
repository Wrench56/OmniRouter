package modmgr

import (
	"context"
	"io/fs"
	"omnirouter/internal/logger"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func LookForChanges(rootpath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("FSNotify setup failed", "error", err)
		return
	}
	defer watcher.Close()

	root := filepath.Clean(rootpath)

	_ = watcher.Add(root)
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			/* Ignore invalid subdirs (permission issues) */
			return fs.SkipDir
		}
		if d != nil {
			if d.IsDir() {
				_ = watcher.Add(path)
			} else {
				CreateModule(path)
			}
		}
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				p := filepath.Clean(event.Name)

				if event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Chmod) {
					fi, err := os.Stat(p)
					if err == nil {
						if fi.IsDir() {
							if err := watcher.Add(p); err == nil {
								logger.Debug("Watch added", "dir", p)
							}
						} else {
							ResetDebounceTimer(p)
						}
					}
				}

				if event.Has(fsnotify.Write) {
					st, err := os.Stat(event.Name)
					if err != nil {
						logger.Warn("stat() returned error while checking file entry WRITE event")
						continue
					}

					if !st.IsDir() {
						ResetDebounceTimer(p)
						logger.Debug("File modified", "path", p)
					}
				}

				if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					RemoveModule(p)
					err = watcher.Remove(p)
					if err == nil {
						logger.Debug("Watch removed", "path", p)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error("Error from FSNotify watcher", "error", err)

			case <-ctx.Done():
				return
			}
		}
	}()

	<-make(chan struct{})
}
