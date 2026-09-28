package model

import "encoding/json"

type SyncData[K comparable, V any] map[K]V

func (d SyncData[K, V]) Get(key K) (v V) {
	if d.Has(key) {
		v = d[key]
	}
	return
}

func (d SyncData[K, V]) Has(key K) bool {
	_, has := d[key]
	return has
}

func (d SyncData[K, V]) Put(key K, value V) {
	d[key] = value
}

func (d SyncData[K, V]) Del(key K) {
	delete(d, key)
}

func (d SyncData[K, V]) Len() int {
	return len(d)
}

func (d SyncData[K, V]) Empty() bool {
	return d.Len() == 0
}

func (d SyncData[K, V]) JSON() []byte {
	out, _ := json.MarshalIndent(d, "", "  ")
	return out
}

func (d SyncData[K, V]) String() string {
	return string(d.JSON())
}
