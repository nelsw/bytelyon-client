package model

import (
	"cmp"
	"maps"
	"slices"
)

type Set[T cmp.Ordered] struct {
	z Map[T, bool]
}

func MakeSet[K cmp.Ordered]() Set[K] {
	return Set[K]{
		z: make(Map[K, bool]),
	}
}

func (s *Set[K]) Add(key K) bool {
	if s.z.Has(key) {
		return false
	}
	s.z.Put(key, true)
	return true
}

func (s *Set[K]) Has(key K) bool {
	return s.z.Has(key)
}

func (s *Set[K]) Keys() []K {
	return slices.Sorted(maps.Keys(s.z))
}

func (s *Set[K]) Len() int {
	return s.z.Len()
}

func (s *Set[K]) Empty() bool {
	return s.z.Empty()
}
