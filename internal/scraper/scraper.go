package scraper

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/deside01/tg_freelance_bot/internal/config"
	"github.com/deside01/tg_freelance_bot/internal/database"
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

func GetOrders() (rssData RssFeed, err error) {
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

	for _, v := range rssData.Channel.Items {
		pubAt, _ := time.Parse(time.RFC1123, v.PublishDate)

		_, err := config.DB.CreateOrder(context.Background(), database.CreateOrderParams{
			Title: v.Title,
			Description: v.Description,
			Link: v.Link,
			PublishedAt: pubAt,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		})

		if err != nil {
			if strings.Contains(err.Error(), "constraint failed: UNIQUE constraint failed: orders.link (2067)") {
				continue
			}

			log.Println("ошибка", err)
		}
	}

	return rssData, nil
}

func GetOrders2() (filteredData []database.Order, err error) {
	var rssData RssFeed
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(FL_URL)
	if err != nil {
		return filteredData, fmt.Errorf("get err: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return filteredData, fmt.Errorf("read err: %v", err)
	}

	err = xml.Unmarshal(data, &rssData)
	if err != nil {
		return filteredData, fmt.Errorf("unmarshal err: %v", err)
	}

	for _, v := range rssData.Channel.Items {
		pubAt, _ := time.Parse(time.RFC1123, v.PublishDate)

		order, err := config.DB.CreateOrder(context.Background(), database.CreateOrderParams{
			Title: v.Title,
			Description: v.Description,
			Link: v.Link,
			PublishedAt: pubAt,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		})

		if err != nil {
			if strings.Contains(err.Error(), "constraint failed: UNIQUE constraint failed: orders.link (2067)") {
				continue
			}

			log.Println("ошибка", err)
			continue
		}

		filteredData = append(filteredData, order)
	}

	return filteredData, nil
}
// func StartScraper(requestDelay time.Duration) {
// 	ticker := time.NewTicker(requestDelay)
// 	for ; ; <-ticker.C {
// 		scraper()
// 	}
// }
