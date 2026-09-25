package model

type Entry[K comparable, V any] struct {
	Key K
	Val V
}

func NewEntry[K comparable, V any](key K, val V) *Entry[K, V] {
	return &Entry[K, V]{Key: key, Val: val}
}
