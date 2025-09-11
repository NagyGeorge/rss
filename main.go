package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide an RSS feed")
		os.Exit(1)
	}

	rssFeed := os.Args[1]
	fmt.Printf("Looking up activity for: %s\n\n", rssFeed)
}
