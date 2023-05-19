package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"

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

const NO = "NO"
const YES = "YES"

func Search(arg_query string, arg_quality string, arg_genre string, arg_rating string, arg_order string, arg_year string, arg_language string) (error, []movie) {
	// Access the webpage
	res, err := http.Get(fmt.Sprintf("https://yts.mx/browse-movies/%s/%s/%s/%s/%s/%s/%s", arg_query, arg_quality, arg_genre, arg_rating, arg_order, arg_year, arg_language))
	//fmt.Printf("https://yts.mx/browse-movies/%s/%s/%s/%s/%s/%s/%s\n\n", arg_query, arg_quality, arg_genre, arg_rating, arg_order, arg_year, arg_language)
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
		} else {
			genres = append(genres, []string{})
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

func Get_links(url string) (error, []download_link) {
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

	fmt.Println("returned")
	return nil, ret
}

func Download(download download_link) (error, []byte) {
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

func YTS_API(communication chan []movie, indexing chan int, w *app.Window) {
	err, movies := Search("avengers", "all", "all", "0", "latest", "0", "en")
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < len(movies); i++ {
		time.Sleep(time.Second / 2)

		communication <- movies
		indexing <- i

		w.Invalidate()
	}

	for {
	}
}

func makeListner(port string) net.Listener {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return nil
	}
	return ln
}

func wireless_API(communication chan []movie, indexing chan int, w *app.Window, sock net.Listener) {
	// Listen in paralel
	var remote_str chan []byte
	remote_str = make(chan []byte, 1)
	go func() {
		for {
			conn, err := sock.Accept()
			if err != nil {
				fmt.Println("Catastrophic error! Please retry.")
				continue
			}
			fmt.Println("Connected!")

			buf := make([]byte, 1024)
			_, err = conn.Read(buf)
			if err != nil {
				fmt.Println("Error reading message:", err.Error())
			}

			remote_str <- buf
			fmt.Println("Passed down data")
		}
	}()

	// Get some default movies
	err, movies := Search(" ", "all", "all", "0", "latest", "0", "en")
	if err != nil {
		fmt.Println(err)
	}
	index := 0

	// Send data
	for {
		select {
		case buf := <-remote_str:
			fmt.Println("Got the data!")

			res := strings.Split(string(buf), "^")
			i := res[0]
			t := res[1]

			err, new_movies := Search(t, "all", "all", "0", "latest", "0", "en")
			if err != nil {
				fmt.Println(err)
			}

			i_, err := strconv.Atoi(i)
			if err != nil {
				fmt.Println(err)
			}

			movies = new_movies
			index = i_

			fmt.Println("Sending...")
			communication <- movies
			indexing <- index
			w.Invalidate()
			fmt.Println("New movie!")

		default:
			communication <- movies
			indexing <- index
		}

	}
}

type C = layout.Context
type D = layout.Dimensions

/*
var movie_list chan []movie
var movie_index chan int
*/
var response []string
var query = ""
var movies []movie
var index = 0

func main() {
	// Default values
	//movie_list = make(chan []movie, 1)
	//movie_index = make(chan int, 1)
	response = []string{" ", "1", "NO"}

	// Set up API
	sock := makeListner("8080")
	go func() {
		for {
			// Connect clients
			conn, err := sock.Accept()
			if err != nil {
				fmt.Println("Catastrophic error! Please retry.")
				continue
			}
			fmt.Println("Connected!")

			// Read data
			buf := make([]byte, 1024)
			_, err = conn.Read(buf)
			if err != nil {
				fmt.Println("Error reading message:", err.Error())
				conn.Close()
				continue
			}

			// send data
			response = strings.Split(string(buf), "^")
			fmt.Println(response[1])
			// Cleanup before convert
			cleanStr := strings.ReplaceAll(response[1], "\x00", "")
			indexed, err := strconv.Atoi(cleanStr)
			if err != nil {
				fmt.Println("Invalid response!")
				fmt.Println(err)
			} else {
				index = indexed
			}
		}
	}()

	go func() {
		w := app.NewWindow(
			app.Title("StreamMaxx - by PetardaAtomica"),
			//app.Fullscreen.Option(),
			//app.Maximized.Option(),
		)
		err := run(w)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			// Make background
			layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(
					func(gtx C) D {
						square := clip.Rect{
							Min: image.Pt(0, 0),
							Max: image.Pt(e.Size.X, e.Size.Y),
						}.Op()
						color := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
						paint.FillShape(gtx.Ops, color, square)
						d := image.Point{Y: 400}
						return layout.Dimensions{Size: d}
					},
				),
			)

			title := material.H1(th, "Welcome to StreamMaxx!\nThis is a project by PetardaAtomica")
			title.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			title.Alignment = text.Middle
			title.TextSize = 24

			// Fetch movies
			if response[0] != query {
				query = response[0]
				var err error
				err, movies = Search(query, "all", "all", "0", "latest", "0", "en")
				if err != nil {
					fmt.Println(err)
				}
			}

			title.Text = movies[index].title

			w.Invalidate()
			title.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func clearChannel(ch chan []movie) {
	for len(ch) > 0 {
		<-ch
	}
}
