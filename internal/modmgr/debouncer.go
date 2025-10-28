package modmgr

import (
	"path/filepath"
	"sync"
	"time"
)

const debounceDelay = 200 * time.Millisecond

var (
	debMu      sync.Mutex
	path2timer = make(map[string]*time.Timer)
)

func ResetDebounceTimer(path string) {
	key, _ := filepath.Abs(path)

	debMu.Lock()
	timer := path2timer[key]
	if timer == nil {
		tmr := time.NewTimer(debounceDelay)
		path2timer[key] = tmr
		debMu.Unlock()

		go func(k string, tm *time.Timer) {
			<-tm.C
			ReloadModule(k)
			debMu.Lock()
			delete(path2timer, k)
			debMu.Unlock()
		}(key, tmr)
		return
	}
	timer.Stop()
	timer.Reset(debounceDelay)
	debMu.Unlock()
}
