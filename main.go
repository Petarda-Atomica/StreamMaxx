package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"

	"github.com/PuerkitoBio/goquery"
	"github.com/micmonay/keybd_event"
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

const ALLOW4K = false
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

	//fmt.Println("returned")
	return nil, ret
}

func Download(download download_link) error {
	url := download.link

	// Send GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the response body into a variable
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("playing_now.torrent", []byte(body), 0644)
	if err != nil {
		return err
	}

	return nil
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
	response = []string{" ", "0", "0", ""}

	go func() {
		w := app.NewWindow(
			app.Title("StreamMaxx - by PetardaAtomica"),
			app.Fullscreen.Option(),
			//app.Maximized.Option(),
		)

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

				go func() {
					for {
						// Read data
						buf := make([]byte, 1024)
						_, err = conn.Read(buf)
						if err != nil {
							fmt.Println("Error reading message:", err.Error())
							return
						}

						// send data
						response = strings.Split(string(buf), "^")
						// Cleanup before convert
						cleanStr := strings.ReplaceAll(response[1], "\x00", "")
						indexed, err := strconv.Atoi(cleanStr)
						if err != nil {
							fmt.Println("Invalid response!")
							fmt.Println(err)
						} else {
							index = indexed
						}

						w.Invalidate()

					}
				}()
			}
		}()

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
				if len(movies) == 0 {
					movies = []movie{{
						title:     "Not Found",
						link:      "https://google.com",
						year:      0,
						rating:    "-1/10",
						genres:    []string{},
						bannerURL: "https://iili.io/HrwBikx.jpg",
					}}
				}
			}

			// Correct indexes
			index = index % len(movies)

			// Build banners
			var banners []image.Image
			banner_distance := 40
			for i := 0; i < len(movies); i++ {
				img, err := fetchImage(movies[i].bannerURL)
				if err != nil {
					fmt.Println(err)
				}
				banners = append(banners, img)

				layout.Stack{}.Layout(gtx,
					layout.Stacked(func(gtx layout.Context) layout.Dimensions {
						var imageOp paint.ImageOp
						if response[0] != query || true {
							// Add your custom object to the ops
							imageOp = paint.NewImageOp(img)
							imageOp.Add(&ops)

							// Apply the desired transformation to the object
							banner_size := f32.Pt(1.05, 1.05)
							op.Affine(f32.Affine2D{}.Scale(f32.Pt(0, 0), banner_size)).Add(&ops)
						}
						// Apply the desired offset to the object
						banner_Y := e.Size.Y / 2
						canvas_center := (e.Size.X - imageOp.Size().X/2) / 2
						if index == i {
							banner_Y += banner_distance
						}
						op.Offset(image.Pt(canvas_center+(i-index)*(banner_distance+imageOp.Size().X), banner_Y)).Add(&ops)

						// Paint the object
						paint.PaintOp{}.Add(&ops)

						// Return the dimensions of the object
						return layout.Dimensions{
							Size: imageOp.Size(),
						}
					}),
				)

			}

			// Cleanup response
			cleanStr := response[2]
			cleanStr = strings.ReplaceAll(response[2], "\x00", "")
			choice, err := strconv.Atoi(cleanStr)
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Printf("Val: %d\nType: %d", choice, reflect.TypeOf(choice))

			// Check if movie should start
			if choice == 1 {
				fmt.Println("Started download")

				// Get download links
				err, available_qualities := Get_links(movies[index].link)
				if err != nil {
					fmt.Println(err)
				}

				// Download movie
				var link download_link
				for i := 1; ; i++ {
					current_quality := strings.Split(available_qualities[len(available_qualities)-i].quality, ".")[0]
					real_quality, err := strconv.Atoi(strings.Replace(current_quality, "p", "", -1))
					if err == nil && (real_quality < 2160 || ALLOW4K) {
						link = available_qualities[len(available_qualities)-i]
						break
					} else {
						//fmt.Println("Wrong codec...")
					}

					// If no quality is found, reverse action
					if i == len(available_qualities)-1 {
						choice = 0
						fmt.Println("!!! No qualities found !!!")
						break
					}
					//fmt.Println("I am stuck in a loop! Help!")

				}
				err = Download(link)
				if err != nil {
					fmt.Println(err)
				}

				// Check OS and play video
				cmd := exec.Command("peerflix", "playing_now.torrent", "--vlc")

				go cmd.Run()
			} else {
				cmd := exec.Command("taskkill", "/F", "/IM", "vlc.exe")

				err := cmd.Run()
				if err != nil {
					fmt.Println(err)
				}
			}

			// Handle key presses
			if response[3] != "" {
				pressKey(response[3])
			}

			title.Text = movies[index].title

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

func fetchImage(url string) (image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	img, err := jpeg.Decode(response.Body)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func pressKey(symbol string) {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		fmt.Println(err)
	}

	// For linux, it is very important to wait 2 seconds
	/*
		if runtime.GOOS == "linux" {
			time.Sleep(2 * time.Second)
		}
	*/

	// Cleanup
	cleanStr := symbol
	cleanStr = strings.ReplaceAll(symbol, "\x00", "")

	// Select keys to be pressed
	press := keybd_event.VK_A
	switch cleanStr {
	case "+":
		press = keybd_event.VK_SPACE
	case ">":
		press = keybd_event.VK_RIGHT
	case "<":
		press = keybd_event.VK_LEFT
	}
	kb.SetKeys(press)

	// Press the selected keys
	err = kb.Launching()
	if err != nil {
		fmt.Println(err)
	}

}
