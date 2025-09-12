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

func parseFeed(xmlContent []byte) {
	var ourFeed RSS
	xml.Unmarshal(xmlContent, &ourFeed)

	fmt.Printf("%s\n\n", ourFeed.Channel.Title)

	for _, v := range ourFeed.Channel.Items {
		fmt.Println("------------------")
		fmt.Printf("%s\n\n", v.Title)
		fmt.Printf("%s\n", v.Link)
		parsedTime, err := time.Parse(time.RFC1123Z, v.PubDate)
		check(err)
		formattedTime := parsedTime.Format(time.DateTime)
		fmt.Printf("%s\n", formattedTime)
	}
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
