package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type RSS struct {
	Channel Channel `xml:"channel"`
}
type Channel struct {
	Title string `xml:"title"`
	Items []Item `xml:"item"`
}
type Item struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
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

/*
func writeFeed(stringFeed string) {
	file, err := os.Create("/tmp/dat1")
	check(err)

	defer file.Close()

	n3, err := file.WriteString(stringFeed)
	check(err)
	fmt.Printf("wrote %d bytes\n", n3)

	file.Sync()
}


func parseFeed(xmlContent string) {
}
*/

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide an RSS feed")
		os.Exit(1)
	}

	rssFeed := os.Args[1]
	fmt.Printf("Looking up activity for: %s\n\n", rssFeed)

	fmt.Printf("%s", fetchFeed(rssFeed))
}
