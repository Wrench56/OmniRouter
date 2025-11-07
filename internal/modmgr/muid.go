package modmgr

import (
	"crypto/rand"
	"encoding/binary"
	"omnirouter/internal/logger"
	"sync"
)

type MUID uint64

var (
	muidMap sync.Map
)

func InitMUID64Map() {
	muidMap.Store(0, 0)
}

func generateMUID64(module *Module) MUID {
	var b [8]byte
	for i := 0; i < 10; i++ {
		_, err := rand.Read(b[:])
		if err != nil {
			continue
		}
		muid := MUID(binary.BigEndian.Uint64(b[:]))
		_, ok := muidMap.Load(muid)
		if !ok {
			muidMap.Store(muid, module)
			return MUID(muid)
		}
	}

	logger.Error("Couldn't create unique module ID in 10 tries, returning invalid UUID")
	return 0
}
