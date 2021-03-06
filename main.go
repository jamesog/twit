package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	ui "github.com/gizak/termui"
)

func main() {
	api := anaconda.NewTwitterApiWithCredentials(
		os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"),
		os.Getenv("CONSUMER_KEY"), os.Getenv("CONSUMER_SECRET_KEY"))
	startUI(api)
}

func startUI(api *anaconda.TwitterApi) {
	err := ui.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	tweetList := ui.NewList()
	tweetList.BorderLabel = "Tweets"
	tweetList.Height = ui.TermHeight()
	tweetList.Overflow = "wrap"

	ui.Body.AddRows(
		ui.NewRow(ui.NewCol(12, 0, tweetList)),
	)

	ui.Body.Align()
	ui.Render(ui.Body)
	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})

	loadTimeline(tweetList, api)
	go updateTweets(tweetList, api)

	ui.Loop()
}

func loadTimeline(tweetList *ui.List, api *anaconda.TwitterApi) {
	v := url.Values{}
	v.Set("count", string(ui.TermHeight()-2))
	timeline, err := api.GetHomeTimeline(v)
	if err != nil {
		log.Fatal(err)
	}

	tweets := make([]string, ui.TermHeight()-2, ui.TermHeight()-2)
	tweetList.Items = tweets
	ui.Render(ui.Body)

	for t := len(timeline) - 1; t >= 0; t-- {
		for i := len(tweets) - 1; i > 0; i-- {
			tweets[i] = tweets[i-1]
		}
		tweets[0] = formatTweet(timeline[t])
		ui.Render(ui.Body)
	}
}

func updateTweets(tweetList *ui.List, api *anaconda.TwitterApi) {
	v := url.Values{}
	// Fetch extended tweets by default
	v.Set("tweet_mode", "extended")
	// "with" specifies to include tweets from accounts the user follows
	v.Set("with", "true")
	stream := api.UserStream(v)
	defer stream.Stop()

	// Get the existing tweet list from termui.
	// By the magic of Go arrays, the underlying array is modified when
	// updating this slice, so no need to re-add the slice to the termui List.
	tweets := tweetList.Items

	for v := range stream.C {
		// Ignore anything that isn't a tweet
		t, ok := v.(anaconda.Tweet)
		if !ok {
			continue
		}

		// Shift each tweet down one in the list
		for j := len(tweets) - 1; j > 0; j-- {
			tweets[j] = tweets[j-1]
		}

		tweets[0] = formatTweet(t)
		ui.Render(ui.Body)
	}
}

func formatTweet(t anaconda.Tweet) string {
	tm, _ := t.CreatedAtTime()
	ts := tm.Format("15:04")
	tt := t.FullText
	// Unwrap t.co URLs
	tt = unwrapURLs(tt, t)
	// Unwrap media
	tt = unwrapMedia(tt, t)
	tu := t.User.ScreenName

	var ru string
	if t.RetweetedStatus != nil {
		ru = fmt.Sprintf(" (via [%s](fg-magenta))", tu)
		tu = t.RetweetedStatus.User.ScreenName
		tt = t.RetweetedStatus.FullText
		// Unwrap t.co URLs
		tt = unwrapURLs(tt, *t.RetweetedStatus)
		// Unwrap media
		tt = unwrapMedia(tt, *t.RetweetedStatus)
	}

	return fmt.Sprintf("[%s](fg-green) [%s](fg-red)%s: %s", ts, tu, ru, tt)
}

// unwrapURLs unwraps (expands) t.co links to the original.
func unwrapURLs(text string, t anaconda.Tweet) string {
	for _, u := range t.Entities.Urls {
		text = strings.Replace(text, u.Url, u.Expanded_url, -1)
	}
	return text
}

// unwrapMedia unwraps (expands) media (embedded photo, GIF, etc.) links from
// t.co to the original.
func unwrapMedia(text string, t anaconda.Tweet) string {
	for _, m := range t.Entities.Media {
		text = strings.Replace(text, m.Url, m.Expanded_url, -1)
	}
	return text
}
