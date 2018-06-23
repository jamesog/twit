package main

import (
	"testing"
)

// Some small benchmarks so I could test the most efficient way of updating a
// slice.

// Test copying a slice of all but the first element back into itself and
// setting the last element - basically a shift/append, but retaining the
// length of the underlying array.
func BenchmarkCopy(b *testing.B) {
	s := []int{1, 2, 3, 4, 5}
	for i := 0; i < b.N; i++ {
		copy(s, s[1:])
		s[cap(s)-1] = i
	}
}

// Similar to the above, but using append(). Due to the way append() works it
// will occasionally add an extra 3 elements to the array when there is only
// one free element.
func BenchmarkSliceAppend(b *testing.B) {
	s := []int{1, 2, 3, 4, 5}
	for i := 0; i < b.N; i++ {
		s = append(s[1:], i)
	}
}

// This is more like an "unshift" where each element is effectively in reverse
// order and we want to shift everything down one (losing the last element)
// and overwriting the first element.
func BenchmarkSliceLoop(b *testing.B) {
	s := []int{5, 4, 3, 2, 1}
	for i := 0; i < b.N; i++ {
		// fmt.Println(s)
		for j := len(s) - 1; j > 0; j-- {
			s[j] = s[j-1]
		}
		s[0] = i
	}
}
