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

	go updateTweets(tweetList, api)

	ui.Loop()
}

func updateTweets(tweetList *ui.List, api *anaconda.TwitterApi) {
	v := url.Values{}
	// Fetch extended tweets by default
	v.Set("tweet_mode", "extended")
	// "with" specifies to include tweets from accounts the user follows
	v.Set("with", "true")
	stream := api.UserStream(v)
	defer stream.Stop()

	tweets := make([]string, ui.TermHeight()-2, ui.TermHeight()-2)
	tweetList.Items = tweets
	ui.Render(ui.Body)

	for v := range stream.C {
		// Ignore anything that isn't a tweet
		t, ok := v.(anaconda.Tweet)
		if !ok {
			continue
		}

		// Ignore retweets
		// if t.RetweetedStatus != nil {
		// 	continue
		// }

		// Shift each tweet down one in the list
		for j := len(tweets) - 1; j > 0; j-- {
			tweets[j] = tweets[j-1]
		}

		tm, _ := t.CreatedAtTime()
		ts := tm.Format("15:04")
		tt := t.FullText
		// Unwrap t.co URLs
		for _, u := range t.Entities.Urls {
			tt = strings.Replace(tt, u.Url, u.Expanded_url, -1)
		}
		// Unwrap media
		for _, m := range t.Entities.Media {
			tt = strings.Replace(tt, m.Url, m.Expanded_url, -1)
		}
		tweets[0] = fmt.Sprintf("[%s](fg-green) [%s](fg-red): %s", ts, t.User.ScreenName, tt)
		ui.Render(ui.Body)
	}
}
