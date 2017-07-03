package cache

import (
	"code.laria.me/petrific/objects"
	"time"
)

type Cache interface {
	PathUpdated(path string) (mtime time.Time, id objects.ObjectId, ok bool)
	SetPathUpdated(path string, mtime time.Time, id objects.ObjectId)
}

type NopCache struct{}

func (NopCache) PathUpdated(_ string) (_ time.Time, _ objects.ObjectId, ok bool) {
	ok = false
	return
}

func (NopCache) SetPathUpdated(_ string, _ time.Time, _ objects.ObjectId) {}
