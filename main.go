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


func fetchFeed(rssURL string) ([]byte, error) {
	res, err := http.Get(rssURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("response failed with status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func parseFeed(xmlContent []byte) error {
	var ourFeed RSS
	err := xml.Unmarshal(xmlContent, &ourFeed)
	if err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	fmt.Printf("%s\n\n", ourFeed.Channel.Title)

	for _, v := range ourFeed.Channel.Items {
		fmt.Println("------------------")
		fmt.Printf("%s\n\n", v.Title)
		fmt.Printf("%s\n", v.Link)

		parsedTime, err := timeChecker(v.PubDate)
		if err != nil {
			fmt.Printf("Could not parse date: %s\n", v.PubDate)
		} else {
			formattedTime := parsedTime.Format(time.DateTime)
			fmt.Printf("%s\n", formattedTime)
		}
	}
	return nil
}

func timeChecker(ourTime string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339Nano,
		time.RFC3339,
		time.RFC850,
	}
	for _, format := range formats {
		if t, err := time.Parse(format, ourTime); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", ourTime)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide an RSS feed")
		os.Exit(1)
	}

	rssFeed := os.Args[1]
	fmt.Printf("Looking up activity for: %s\n\n", rssFeed)

	xmlContent, err := fetchFeed(rssFeed)
	if err != nil {
		fmt.Printf("Error fetching feed: %v\n", err)
		os.Exit(1)
	}

	if err := parseFeed(xmlContent); err != nil {
		fmt.Printf("Error parsing feed: %v\n", err)
		os.Exit(1)
	}
}
