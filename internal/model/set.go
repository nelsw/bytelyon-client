package model

import (
	"cmp"
	"maps"
	"slices"
	"sync"
)

type Set[K cmp.Ordered] struct {
	x sync.Mutex
	m map[K]bool
}

func NewSet[K cmp.Ordered]() *Set[K] {
	return &Set[K]{
		m: make(map[K]bool),
	}
}

func (s *Set[K]) Put(key K, val bool) bool {
	s.x.Lock()
	defer s.x.Unlock()
	if _, ok := s.m[key]; ok {
		return false
	}
	s.m[key] = val
	return true
}

// Update overwrites the value of an existing key, reporting whether it was present.
// Unlike Put, it does not add the key.
func (s *Set[K]) Update(key K, val bool) bool {
	s.x.Lock()
	defer s.x.Unlock()
	if _, ok := s.m[key]; !ok {
		return false
	}
	s.m[key] = val
	return true
}

func (s *Set[K]) Has(key K) bool {
	s.x.Lock()
	defer s.x.Unlock()
	_, ok := s.m[key]
	return ok
}

func (s *Set[K]) Delete(key K) {
	s.x.Lock()
	defer s.x.Unlock()
	delete(s.m, key)
}

func (s *Set[K]) Keys() []K {
	s.x.Lock()
	defer s.x.Unlock()
	return slices.Sorted(maps.Keys(maps.Clone(s.m)))
}
