package main

import "testing"

func TestClampConns(t *testing.T) {
	cases := map[int]int32{-5: 1, 0: 1, 1: 1, 25: 25, 1000: 1000, 5000: 1000}
	for in, want := range cases {
		if got := clampConns(in); got != want {
			t.Errorf("clampConns(%d) = %d, want %d", in, got, want)
		}
	}
}
