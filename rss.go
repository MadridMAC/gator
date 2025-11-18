package main

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"log"
	"net/http"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// make custom client so header can be edited
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// create RSS GET request w/ proper set header
	rss_request, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		log.Fatalf("error making rss request: %v", err)
	}
	rss_request.Header.Set("User-Agent", "gator")

	// receive response and defer closing the body
	resp, err := client.Do(rss_request)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// read data obtained from the response
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// unmarshal the XML into the RSSFeed struct
	var output *RSSFeed
	if err := xml.Unmarshal(data, &output); err != nil {
		log.Fatalf("error unmarshalling XML: %v", err)
	}

	return output, nil
}

func feedFormatter(feed *RSSFeed) *RSSFeed {
	// decodes escaped HTML entities in the Title and Description fields of the provided RSSFeed and its RSSItems
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i, item := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(item.Title)
		feed.Channel.Item[i].Description = html.UnescapeString(item.Description)
	}
	return feed
}
