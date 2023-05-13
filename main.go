package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type movie struct {
	title     string
	link      string
	year      int
	rating    string
	genres    []string
	bannerURL string
}

type download_link struct {
	quality string
	link    string
}

func search(arg_query string, arg_quality string, arg_genre string, arg_rating string, arg_order string, arg_year string, arg_language string) (error, []movie) {
	// Access the webpage
	res, err := http.Get(fmt.Sprintf("https://yts.mx/browse-movies/%s/%s/%s/%s/%s/%s/%s", arg_query, arg_quality, arg_genre, arg_rating, arg_order, arg_year, arg_language))
	if err != nil {
		return err, nil
	}
	defer res.Body.Close()

	// Load HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err, nil
	}

	// Define variables
	var titles []string
	var links []string
	var years []int
	var ratings []string
	var genres [][]string
	var banners []string

	// Get titles and links
	for i := 1; ; i++ {
		divPath := fmt.Sprintf("body div:nth-of-type(4) div:nth-of-type(4) div section div div:nth-of-type(%d) div", i)
		div := doc.Find(divPath)
		if div.Length() == 0 {
			// No more <div> elements at the specified path
			break
		}

		a := div.Find("a")
		if a.Length() == 0 {
			// No <a> element inside the <div> element
			continue
		}

		// Retrieve the text and href attributes of the <a> element
		title := a.Text()
		link, _ := a.Attr("href")

		// Append to the slice of titles and links
		titles = append(titles, title)
		links = append(links, link)
	}

	// Get years
	for i := 1; ; i++ {
		divPath := fmt.Sprintf("body div:nth-of-type(4) div:nth-of-type(4) div section div div:nth-of-type(%d) div", i)
		div := doc.Find(divPath)
		if div.Length() == 0 {
			// No more <div> elements at the specified path
			break
		}

		// Select the <div> element at div[1]
		div1 := div.Find("div:nth-of-type(1)")
		if div1.Length() == 0 {
			// No <div> element at div[1]
			continue
		}

		// Retrieve the text content of the <div> element
		text := div1.Text()

		// Append to the slice of years
		number, _ := strconv.Atoi(text)
		years = append(years, number)
	}

	// Get years and genres
	for i := 1; ; i++ {
		divPath := fmt.Sprintf("body div:nth-of-type(4) div:nth-of-type(4) div section div div:nth-of-type(%d)", i)
		div := doc.Find(divPath)
		if div.Length() == 0 {
			// No more <div> elements at the specified path
			break
		}

		// Select the <a> element
		a := div.Find("a")
		if a.Length() == 0 {
			// No <a> element
			continue
		}

		// Retrieve the text content of the <h4> element at h4[1]
		h4 := a.Find("figure figcaption h4:nth-of-type(1)")
		if h4.Length() == 0 {
			// No <h4> element at h4[1]
			continue
		}

		// Append the text content to the ratings slice
		text := strings.TrimSpace(h4.Text())
		ratings = append(ratings, text)

		// Retrieve the text content of all the <h4> elements after h4[1]
		h4Elements := a.Find("figure figcaption h4:nth-of-type(n+2)")
		var genreTexts []string
		h4Elements.Each(func(_ int, h4Element *goquery.Selection) {
			text := strings.TrimSpace(h4Element.Text())
			if text != "" {
				genreTexts = append(genreTexts, text)
			}
		})

		// Append the slice of genre texts to the genres slice
		if len(genreTexts) > 0 {
			genres = append(genres, genreTexts)
		}
	}

	// Get banners
	doc.Find("html > body > div:nth-of-type(4) > div:nth-of-type(4) > div > section > div > div > a > figure > img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists {
			banners = append(banners, src)
		}
	})

	// Build response
	var ret []movie
	for i := 0; i < len(titles); i++ {
		_movie := movie{
			title:     titles[i],
			link:      links[i],
			year:      years[i],
			rating:    ratings[i],
			genres:    genres[i],
			bannerURL: banners[i],
		}
		ret = append(ret, _movie)
	}

	return nil, ret
}

func get_links(url string) (error, []download_link) {
	cssSelector := "body div:nth-of-type(4) div:nth-of-type(3) div:nth-of-type(1) div:nth-of-type(4) p a"

	// Send a GET request to the website and load the response body into a GoQuery document
	response, err := http.Get(url)
	if err != nil {
		return err, []download_link{}
	}
	defer response.Body.Close()
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return err, []download_link{}
	}

	// Predefine some variables
	var qualities []string
	var links []string

	// Find all elements that match the CSS selector and print their text and href attributes
	document.Find(cssSelector).Each(func(i int, element *goquery.Selection) {
		qualities = append(qualities, element.Text())
		href, _ := element.Attr("href")
		links = append(links, href)
	})

	// Build return
	var ret []download_link
	for i := 0; i < len(links); i++ {
		_download_link := download_link{
			quality: qualities[i],
			link:    links[i],
		}

		ret = append(ret, _download_link)
	}

	return nil, ret
}

func download(download download_link) (error, []byte) {
	url := download.link

	// Send GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return err, []byte{}
	}
	defer resp.Body.Close()

	// Read the response body into a variable
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, []byte{}
	}

	return nil, body
}
