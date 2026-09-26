package model

import "encoding/json"

type Data[K comparable, V any] map[K]V

func (d Data[K, V]) Get(key K) (v V) {
	if d.Has(key) {
		v = d[key]
	}
	return
}

func (d Data[K, V]) Has(key K) bool {
	_, has := d[key]
	return has
}

func (d Data[K, V]) Put(key K, value V) {
	d[key] = value
}

func (d Data[K, V]) Del(key K) {
	delete(d, key)
}

func (d Data[K, V]) Len() int {
	return len(d)
}

func (d Data[K, V]) Empty() bool {
	return d.Len() == 0
}

func (d Data[K, V]) JSON() []byte {
	out, _ := json.MarshalIndent(d, "", "  ")
	return out
}

func (d Data[K, V]) String() string {
	return string(d.JSON())
}
