package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	"github.com/gocolly/colly"
)

type Music struct {
	ID    int    `json:id`
	Title string `json:title`
	Audio string `json:audio`
	Genre string `json:genre`
	Image string `json:image`
}

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("www.songsio.com"),
	)

	fetchMusic := c.Clone()

	// ch := make(chan map[string]string)

	ch_title := make(chan string)
	ch_link := make(chan string)
	ch_image := make(chan string)

	// array := make(chan []interface{})

	c.OnHTML("article.vce-post > div > a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		title := e.Attr("title")

		if strings.Contains(title, "CAPRISONGS") {
			return
		}

		go func() {
			ch_title <- title
		}()

		fetchMusic.Visit(link)
	})

	fetchMusic.OnHTML("div.list-group > p:nth-child(2) > a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.Contains(link, "http://ecdn31.frkmusic.xyz/caprisongs/") {
			return
		}

		go func() {
			ch_link <- link
		}()
	})

	// gets the music image cover
	fetchMusic.OnHTML("div > p > img", func(e *colly.HTMLElement) {
		image := e.Attr("src")
		go func() {
			ch_image <- image
		}()
	})

	fetchMusic.OnHTML("div > h2:nth-child(6) > span > span > span:nth-child(2)", func(e *colly.HTMLElement) {
		genre := e.Text

		music := Music{
			Title: <-ch_title,
			Genre: genre,
			Audio: <-ch_link,
			Image: <-ch_image,
		}

		file, err := ioutil.ReadFile("output.json")
		if err != nil {
			log.Fatal(err)
		}

		var data []interface{}

		json.Unmarshal(file, &data)
		data = append(data, music)

		dataBytes, _ := json.MarshalIndent(data, "", "\t")
		err = ioutil.WriteFile("output.json", dataBytes, 0644)
		if err != nil {
			log.Fatal(err)
		}
	})

	c.Visit("https://www.songsio.com/genre/all/electronic/")

}
