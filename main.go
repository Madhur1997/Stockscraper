package main

import (
	"context"
	"log"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/chromedp/chromedp"
)

// how many threads to use within the application
const NCPU = 8

// Our crawler structure definition
type Crawler struct {
	urls chan string
}

func (crawler *Crawler) start(wg *sync.WaitGroup) {
	// wait for new URLs to be extracted and passed to the URLs channel.
	wg.Add(1)
	go func() {
		for url := range crawler.urls {
			wg.Add(1)
			go crawler.googleSearch(url, q, wg)
		}
		wg.Done()
	}()
}

func (crawler *Crawler) getContents(url string) {

}

// given a URL, this method retrieves the required element from the web page.
func (crawler *Crawler) googleSearch(url string, wg *sync.WaitGroup) {

	defer wg.Done()

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	sel := "//input[@name='q']"
	q := "reliance stock price"

	// run task list
	var res string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(sel),
		chromedp.SendKeys(sel, q),
		chromedp.Click(`input[name="btnK"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`//span[@jsname="vWLAgc"]`),
		chromedp.Text(`//span[@jsname="vWLAgc"]`, &res),
	)

	if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile("\\n")
	res = re.ReplaceAllString(res, " ")
	log.Println(strings.ToUpper(q) + ": " + res)
}

// stops the crawler by closing both the URLs channel
func (crawler *Crawler) stop() {
	close(crawler.urls)
}

func main() {
	// set how many processes (threads to use)
	runtime.GOMAXPROCS(NCPU)

	var wg sync.WaitGroup
	// create a new instance of the crawler structure
	c := &Crawler{
		make(chan string),
	}
	
	c.start(&wg)

	c.urls <- "https://www.google.com"
	c.stop()

	wg.Wait()

}
