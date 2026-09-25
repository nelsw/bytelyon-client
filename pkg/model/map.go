package model

type Map[K comparable, V any] map[K]V

func (m Map[K, V]) Get(key K) (v V) {
	if m.Has(key) {
		v = m[key]
	}
	return
}

func (m Map[K, V]) Has(key K) bool {
	_, has := m[key]
	return has
}

func (m Map[K, V]) Put(key K, value V) {
	m[key] = value
}

func (m Map[K, V]) Del(key K) {
	delete(m, key)
}

func (m Map[K, V]) Len() int {
	return len(m)
}

func (m Map[K, V]) Empty() bool {
	return m.Len() == 0
}
