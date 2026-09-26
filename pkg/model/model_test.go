package model

import (
	"slices"
	"sync"
	"testing"
)

func TestData(t *testing.T) {
	d := Data[string, any]{}
	if !d.Empty() || d.Len() != 0 {
		t.Fatal("new data should be empty")
	}
	if d.Get("missing") != nil || d.Has("missing") {
		t.Error("missing key should be absent with zero value")
	}

	d.Put("a", 1)
	d.Put("b", "two")
	if d.Empty() || d.Len() != 2 || !d.Has("a") || d.Get("b") != "two" {
		t.Errorf("unexpected data after puts: %v", d)
	}

	if got, want := d.String(), "{\n  \"a\": 1,\n  \"b\": \"two\"\n}"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}

	d.Del("a")
	if d.Has("a") || d.Len() != 1 {
		t.Errorf("Del did not remove key: %v", d)
	}
}

func TestSet(t *testing.T) {
	s := MakeSet[string]()
	if !s.Empty() || s.Len() != 0 {
		t.Fatal("new set should be empty")
	}
	if !s.Add("b") || !s.Add("a") || s.Add("a") {
		t.Error("Add should report whether the key was new")
	}
	if !s.Has("a") || s.Has("c") || s.Empty() || s.Len() != 2 {
		t.Errorf("unexpected set state: %v", s.Keys())
	}
	if got := s.Keys(); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("Keys() = %v, want sorted [a b]", got)
	}
}

func TestSyncSet(t *testing.T) {
	s := NewSyncSet[int]()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() { s.Add(i % 10) })
	}
	wg.Wait()

	if s.Add(3) {
		t.Error("Add should return false for an existing key")
	}
	s.AddAll([]int{10, 11, 3})

	want := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	if got := s.Keys(); !slices.Equal(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
	if !s.Has(11) || s.Has(12) {
		t.Error("Has returned the wrong result")
	}

	v, err := s.Value()
	if err != nil || !slices.Equal(v.([]int), want) {
		t.Errorf("Value() = %v, %v", v, err)
	}
}
