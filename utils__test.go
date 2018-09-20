package main

import (
	// "reflect"
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
