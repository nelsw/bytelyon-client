package model

import (
	"cmp"
	"database/sql/driver"
	"sync"
)

type SyncSet[T cmp.Ordered] struct {
	z Set[T]
	x sync.RWMutex
}

func NewSyncSet[K cmp.Ordered]() *SyncSet[K] {
	return &SyncSet[K]{
		z: MakeSet[K](),
	}
}

func (s *SyncSet[K]) Add(key K) bool {
	s.x.Lock()
	defer s.x.Unlock()
	return s.z.Add(key)
}

func (s *SyncSet[K]) AddAll(keys []K) {
	s.x.Lock()
	defer s.x.Unlock()
	for _, key := range keys {
		s.z.Add(key)
	}
}

func (s *SyncSet[K]) Has(key K) bool {
	s.x.RLock()
	defer s.x.RUnlock()
	return s.z.Has(key)
}

func (s *SyncSet[K]) Keys() []K {
	s.x.RLock()
	defer s.x.RUnlock()
	return s.z.Keys()
}

func (s *SyncSet[K]) Value() (driver.Value, error) {
	return s.Keys(), nil
}
