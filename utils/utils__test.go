package utils

import (
	"reflect"
	"testing"
)

func TestIndexString__unsorted_positive(t *testing.T) {
	aa := []string{"eth3", "eth2eeee", "eth5", "eth1", "eth0"}
	if IndexString(aa, "eth2eeee") != 1 {
		t.Fail()
	}
}
func TestIndexString__unsorted_double_positive(t *testing.T) {
	aa := []string{"eth3", "eth2", "eth5", "eth2", "eth1", "eth0"}
	if IndexString(aa, "eth2") != 1 {
		t.Fail()
	}
}
func TestIndexString__unsorted_negative(t *testing.T) {
	aa := []string{"eth3", "eth2", "eth5", "eth1", "eth0"}
	if IndexString(aa, "XXX") >= 0 {
		t.Fail()
	}
}

func TestReverseString(t *testing.T) {
	aa := []string{"a1", "a2", "a3", "a4", "a5"}
	rv := ReverseString(aa)
	if !reflect.DeepEqual(rv, []string{"a5", "a4", "a3", "a2", "a1"}) {
		t.Fail()
	}
}

func TestPrepndString(t *testing.T) {
	aa := []string{"eth1", "eth2eeee", "eth3", "eth5"}
	rv := PrependString(aa, "eth0")
	if !reflect.DeepEqual(rv, []string{"eth0", "eth1", "eth2eeee", "eth3", "eth5"}) {
		t.Fail()
	}
}
