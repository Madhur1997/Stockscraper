package main

import (
	"context"
	"log"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// how many threads to use within the application
const NCPU = 8

// Our crawler structure definition
type Crawler struct {
}

func (crawler *Crawler) start(wg *sync.WaitGroup, queries ...string) {

	log.Println("Starting Web Crawler")

	url := "https://www.google.com"

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, q := range queries {
			wg.Add(1)

			go crawler.scrapStockPrice(url, q, wg)
		}
	}()
}

// given a URL, this method retrieves the required element from the web page.
func (crawler *Crawler) scrapStockPrice(url, q string, wg *sync.WaitGroup) {

	log.Printf("Scrapping stock price for: %s\n", strings.Title(q))

	defer wg.Done()

	inQ := q + " stock price"
	inTextSel := `//input[@name='q']`
	btnSel := `input[name="btnK"]`
	outTextSel := `//span[@jsname="vWLAgc"]`

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())

	// Wait for timeout.
	timeoutContext, cancel := context.WithTimeout(ctx, 10 * time.Second)
	defer cancel()

	// run task list
	var res string
	err := chromedp.Run(
		timeoutContext,
		chromedp.Navigate(url),
		chromedp.WaitVisible(inTextSel),
		chromedp.SendKeys(inTextSel, inQ),
		chromedp.Click(btnSel, chromedp.ByQuery),
		chromedp.WaitVisible(outTextSel),
		chromedp.Text(outTextSel, &res),
	)

	if err != nil {
		log.Printf("Error while scrapping stock price for %s: %v", strings.Title(q), err)
		return
	}

	re := regexp.MustCompile("\\n")
	res = re.ReplaceAllString(res, " ")
	log.Println(strings.ToUpper(q) + ": " + res)
}

func main() {
	// set how many processes (threads to use)
	runtime.GOMAXPROCS(NCPU)

	var wg sync.WaitGroup
	// create a new instance of the crawler structure
	c := &Crawler{
	}
	
	c.start(&wg, "reliance", "ashok leyland", "indigo", "kesoram", "hdfc bank", "Bank Nifty")

	wg.Wait()
}
