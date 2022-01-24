package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/gocolly/colly"
)

type Music struct {
	// ID    int    `json:id`
	Title string `json:title`
	Audio string `json:audio`
	Genre string `json:genre`
	// Image string `json:image`
}

func renderJson(rw http.ResponseWriter, r *http.Request) {
	dataByte, _ := ioutil.ReadFile("output.json")
	var music *Music
	json.Unmarshal(dataByte, &music)
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(dataByte)
}

func main() {

	http.HandleFunc("/", renderJson)
	log.Fatal(http.ListenAndServe(":8000", nil))

	c := colly.NewCollector(
		colly.AllowedDomains("www.songsio.com"),
	)

	fetchMusic := c.Clone()

	ch_title := make(chan string)
	ch_link := make(chan string)
	// ch_image := make(chan string)

	c.OnHTML("article.vce-post > div > a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fetchMusic.Visit(link)
	})

	fetchMusic.OnHTML("div.list-group > p:nth-child(2) > a", func(e *colly.HTMLElement) {

		pattern := regexp.MustCompile(`^[^#\n].*`)
		title := pattern.FindAllString(e.Text, -1)[0]
		link := e.Attr("href")

		go func() {
			ch_link <- link
		}()

		go func() {
			ch_title <- title
		}()

	})

	// gets the music image cover
	// fetchMusic.OnHTML("div > p > img", func(e *colly.HTMLElement) {
	// 	image := e.Attr("src")
	// 	go func() {
	// 		ch_image <- image
	// 	}()
	// })

	fetchMusic.OnHTML("div > h2:nth-child(6) > span > span > span:nth-child(2)", func(e *colly.HTMLElement) {
		genre := e.Text

		music := &Music{
			Title: <-ch_title,
			Genre: genre,
			Audio: <-ch_link,
			// Image: <-ch_image,
		}

		file, err := ioutil.ReadFile("output.json")
		if err != nil {
			log.Fatal(err)
		}

		// var data []interface{}
		var data []*Music

		json.Unmarshal(file, &data)
		data = append(data, music)

		dataBytes, _ := json.MarshalIndent(data, "", "\t")
		err = ioutil.WriteFile("output.json", dataBytes, 0644)
		if err != nil {
			log.Fatal(err)
		}
	})

	// c.Visit("https://www.songsio.com/genre/all/electronic/")
	for i := 1; i < 6; i++ {
		c.Visit(fmt.Sprintf("https://www.songsio.com/genre/all/electronic/page/%d", i))
	}

}
