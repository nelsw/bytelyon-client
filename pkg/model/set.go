package model

import (
	"cmp"
	"maps"
	"slices"
	"sync"
)

type Set[K cmp.Ordered] struct {
	x     sync.Mutex
	m     map[K]bool
	async bool
}

func NewSet[K cmp.Ordered](async ...bool) *Set[K] {
	return &Set[K]{
		m:     make(map[K]bool),
		async: len(async) > 0 && async[0],
	}
}

func (s *Set[K]) Put(key K, val bool) bool {
	if !s.async {
		s.x.Lock()
		defer s.x.Unlock()
	}
	if _, ok := s.m[key]; ok {
		return false
	}
	s.m[key] = val
	return true
}

// Update overwrites the value of an existing key, reporting whether it was present.
// Unlike Put, it does not add the key.
func (s *Set[K]) Update(key K, val bool) bool {
	if !s.async {
		s.x.Lock()
		defer s.x.Unlock()
	}
	if _, ok := s.m[key]; !ok {
		return false
	}
	s.m[key] = val
	return true
}

func (s *Set[K]) Has(key ...K) bool {
	if !s.async {
		s.x.Lock()
		defer s.x.Unlock()
	}
	for _, k := range key {
		if _, ok := s.m[k]; !ok {
			return false
		}
	}
	return true
}

func (s *Set[K]) Delete(key K) {
	if !s.async {
		s.x.Lock()
		defer s.x.Unlock()
	}
	delete(s.m, key)
}

func (s *Set[K]) Keys() []K {
	if !s.async {
		s.x.Lock()
		defer s.x.Unlock()
	}
	return slices.Sorted(maps.Keys(maps.Clone(s.m)))
}
