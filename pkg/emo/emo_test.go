package emo

import "testing"

func TestBool(t *testing.T) {
	if Bool(true) != Truthy || Bool(false) != Falsey {
		t.Errorf("Bool() = %q, %q", Bool(true), Bool(false))
	}
}
