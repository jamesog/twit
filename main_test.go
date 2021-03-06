package main

import (
	"testing"

	"github.com/ChimeraCoder/anaconda"
)

var (
	testTweet = anaconda.Tweet{
		CreatedAt: "Sat Feb 07 19:38:05 +0000 2009",
		FullText:  "Test tweet https://t.co/test1 https://t.co/test2",
		User: anaconda.User{
			Name:       "James O'Gorman",
			ScreenName: "jogbert",
		},
		Entities: anaconda.Entities{
			Urls: []struct {
				Indices      []int  `json:"indices"`
				Url          string `json:"url"`
				Display_url  string `json:"display_url"`
				Expanded_url string `json:"expanded_url"`
			}{
				{Url: "https://t.co/test1", Expanded_url: "https://example.com/test/1"},
				{Url: "https://t.co/test2", Expanded_url: "https://example.com/test/2"},
			},
			Media: []anaconda.EntityMedia{
				{Url: "https://t.co/test1", Expanded_url: "https://example.com/test/1"},
				{Url: "https://t.co/test2", Expanded_url: "https://example.com/test/2"},
			},
		},
	}

	testRetweet = anaconda.Tweet{
		CreatedAt: "Sat Feb 07 20:38:05 +0000 2009",
		FullText:  "RT @jogbert: Test tweet https://t.co/test1 https://t.co/test2",
		User: anaconda.User{
			Name:       "Ryan Carter",
			ScreenName: "vaelen",
		},
		RetweetedStatus: &testTweet,
	}
)

func TestFormatTweet(t *testing.T) {
	t.Run("Tweet", func(t *testing.T) {
		want := "[19:38](fg-green) [jogbert](fg-red): Test tweet https://example.com/test/1 https://example.com/test/2"
		got := formatTweet(testTweet)
		if want != got {
			t.Errorf("want %q, got %q", want, got)
		}
	})

	t.Run("Retweet", func(t *testing.T) {
		want := "[20:38](fg-green) [jogbert](fg-red) (via [vaelen](fg-magenta)): Test tweet https://example.com/test/1 https://example.com/test/2"
		got := formatTweet(testRetweet)
		if want != got {
			t.Errorf("want %q, got %q", want, got)
		}
	})

}

func TestUnwrapURLs(t *testing.T) {
	want := "Test tweet https://example.com/test/1 https://example.com/test/2"
	got := unwrapURLs(testTweet.FullText, testTweet)
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestUnwrapMedia(t *testing.T) {
	want := "Test tweet https://example.com/test/1 https://example.com/test/2"
	got := unwrapMedia(testTweet.FullText, testTweet)
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

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
