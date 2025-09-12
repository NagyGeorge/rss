package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type RSS struct {
	Channel Channel `xml:"channel"`
}
type Channel struct {
	Title string `xml:"title"`
	Items []Item `xml:"item"`
}
type Item struct {
	Title    string `xml:"title"`
	Link     string `xml:"link"`
	Comments string `xml:"comments"`
	PubDate  string `xml:"pubDate"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func fetchFeed(rssURL string) []byte {
	res, err := http.Get(rssURL)
	check(err)
	defer res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Printf("Response failed with status code: %d\n", res.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(res.Body)
	check(err)

	return body
}

func parseFeed(xmlContent []byte) error {
	var ourFeed RSS
	err := xml.Unmarshal(xmlContent, &ourFeed)
	check(err)

	fmt.Printf("%s\n\n", ourFeed.Channel.Title)

	for _, v := range ourFeed.Channel.Items {
		fmt.Println("------------------")
		fmt.Printf("%s\n\n", v.Title)
		fmt.Printf("%s\n", v.Link)

		parsedTime := timeChecker(v.PubDate)
		formattedTime := parsedTime.Format(time.DateTime)
		fmt.Printf("%s\n", formattedTime)
	}
	return nil
}

func timeChecker(ourTime string) time.Time {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339Nano,
		time.RFC3339,
		time.RFC850,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
		time.DateTime,
		time.DateOnly,
	}
	var goodTime time.Time
	for _, v := range formats {
		if t, err := time.Parse(v, ourTime); err == nil {
			goodTime = t
		}
	}
	return goodTime
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide an RSS feed")
		os.Exit(1)
	}

	rssFeed := os.Args[1]
	fmt.Printf("Looking up activity for: %s\n\n", rssFeed)

	parseFeed(fetchFeed(rssFeed))
}
