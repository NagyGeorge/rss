package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	const maxRetries = 3
	const timeout = 10 * time.Second
	const maxSize = 10 << 20 // 10MB

	client := &http.Client{Timeout: timeout}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		req, err := http.NewRequestWithContext(ctx, "GET", rssURL, nil)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create request for %s: %w", rssURL, err)
		}
		req.Header.Set("User-Agent", "RSS-Reader/1.0")

		res, err := client.Do(req)
		if err != nil {
			cancel()
			lastErr = fmt.Errorf("network error on attempt %d/%d: %w", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return nil, lastErr
		}

		if res.StatusCode < 200 || res.StatusCode >= 300 {
			res.Body.Close()
			cancel()
			lastErr = fmt.Errorf("server returned %d %s for %s", res.StatusCode, http.StatusText(res.StatusCode), rssURL)
			if attempt < maxRetries && res.StatusCode >= 500 {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return nil, lastErr
		}

		contentType := res.Header.Get("Content-Type")
		if !strings.Contains(strings.ToLower(contentType), "xml") && !strings.Contains(strings.ToLower(contentType), "rss") {
			res.Body.Close()
			cancel()
			return nil, fmt.Errorf("unexpected content type '%s' (expected XML/RSS) from %s", contentType, rssURL)
		}

		body, err := io.ReadAll(io.LimitReader(res.Body, maxSize))
		res.Body.Close()
		cancel()

		if err != nil {
			lastErr = fmt.Errorf("failed to read response body on attempt %d/%d: %w", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return nil, lastErr
		}

		return body, nil
	}

	return nil, lastErr
}

func parseFeed(xmlContent []byte) error {
	var ourFeed RSS
	err := xml.Unmarshal(xmlContent, &ourFeed)
	if err != nil {
		return fmt.Errorf("failed to parse XML content (check if URL returns valid RSS/XML): %w", err)
	}

	if ourFeed.Channel.Title == "" {
		fmt.Println("Warning: Feed has no title")
	} else {
		fmt.Printf("%s\n\n", ourFeed.Channel.Title)
	}

	if len(ourFeed.Channel.Items) == 0 {
		fmt.Println("No items found in this RSS feed")
		return nil
	}

	successCount := 0
	for i, v := range ourFeed.Channel.Items {
		if v.Title == "" && v.Link == "" {
			fmt.Printf("Skipping item %d: missing both title and link\n", i+1)
			continue
		}

		fmt.Println("------------------")

		if v.Title == "" {
			fmt.Printf("(No title)\n\n")
		} else {
			fmt.Printf("%s\n\n", v.Title)
		}

		if v.Link == "" {
			fmt.Printf("(No link available)\n")
		} else {
			fmt.Printf("%s\n", v.Link)
		}

		if v.PubDate != "" {
			parsedTime, err := timeChecker(v.PubDate)
			if err != nil {
				fmt.Printf("Could not parse date '%s': %v\n", v.PubDate, err)
			} else {
				formattedTime := parsedTime.Format(time.DateTime)
				fmt.Printf("%s\n", formattedTime)
			}
		}
		successCount++
	}

	if successCount < len(ourFeed.Channel.Items) {
		fmt.Printf("\nProcessed %d/%d items successfully\n", successCount, len(ourFeed.Channel.Items))
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
		fmt.Println("Usage: rss <feed-url>")
		fmt.Println("Example: rss https://feeds.bbci.co.uk/news/rss.xml")
		os.Exit(1)
	}

	rssFeed := os.Args[1]
	fmt.Printf("Fetching RSS feed from: %s\n\n", rssFeed)

	xmlContent, err := fetchFeed(rssFeed)
	if err != nil {
		fmt.Printf(" Failed to fetch feed: %v\n", err)
		fmt.Println("• Check if the URL is correct and accessible")
		fmt.Println("• Verify you have internet connectivity")
		fmt.Println("• Try the URL in a web browser first")
		os.Exit(1)
	}

	if err := parseFeed(xmlContent); err != nil {
		fmt.Printf(" Failed to parse feed: %v\n", err)
		fmt.Println("• Ensure the URL points to a valid RSS/XML feed")
		fmt.Println("• Some websites require specific user agents or headers")
		os.Exit(1)
	}

	fmt.Println("\n Feed processed successfully")
}
