package scraper

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RssFeed struct {
	Channel struct {
		Items []RssItem `xml:"item"`
	} `xml:"channel"`
}

type RssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PublishDate string `xml:"pubDate"`
}

const FL_URL = "https://www.fl.ru/rss/all.xml?category=5"

func scraper() (rssData RssFeed, err error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(FL_URL)
	if err != nil {
		return rssData, fmt.Errorf("get err: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return rssData, fmt.Errorf("read err: %v", err)
	}

	err = xml.Unmarshal(data, &rssData)
	if err != nil {
				return rssData, fmt.Errorf("unmarshal err: %v", err)
	}

	// for _, v := range rssData.Channel.Items {
		fmt.Println(len(rssData.Channel.Items))
	// }

	return rssData, nil
}

func StartScraper(requestDelay time.Duration) {
	ticker := time.NewTicker(requestDelay)
	for ; ; <-ticker.C {
		scraper()
	}
}