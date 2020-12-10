package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/urfave/cli/v2"
)

// how many threads to use within the application
const NCPU = 1

var personalList []string = []string{"reliance", "ashok leyland", "indigo", "kesoram", "hdfc bank", "adani green energy",
	"vodafone idea", "TCS", "divis labs",}

// Our crawler structure definition
type Crawler struct {
}

func fetchPriceFromGoogle(q string) (string) {

	url := "https://www.google.com"
	inQ := q + " stock price"
	inTextSel := `//input[@name='q']`
	btnSel := `input[name="btnK"]`
	outTextSel := `//span[@jsname="vWLAgc"]`

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	// Wait for timeout.
	timeoutContext, _ := context.WithTimeout(ctx, 30 * time.Second)

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
		chromedp.FromContext(ctx).Allocator.Wait()
		return ""
	}

	re := regexp.MustCompile("\\n")
	return re.ReplaceAllString(res, " ")
}

func (crawler *Crawler) scrapStockPrice(wg *sync.WaitGroup, q string) {
	log.Printf("Scrapping stock price for: %s\n", strings.Title(q))

	defer wg.Done()
	res := fetchPriceFromGoogle(q)
	if res != "" {
		log.Println(strings.ToUpper(q) + ": " + res)
	}
}

func (crawler *Crawler) monitor(wg *sync.WaitGroup, q string) {
	log.Println("Monitoring " + q)

	ticker := time.NewTicker(60 * time.Second)

	for {
		select {
			case <-ticker.C:
				res := fetchPriceFromGoogle(q)
				log.Println(strings.ToUpper(q) + ": " + res)
			// case <-exit:
			// 	log.Println("Received exit notification. Stop monitoring " + q + " and return")
			// 	break
		}
	}
	wg.Done()
}

func (crawler *Crawler) start(wg *sync.WaitGroup, queries ...string) {
	log.Println("Starting Web Crawler")

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, q := range queries {
			wg.Add(1)

			go crawler.scrapStockPrice(wg, q)
		}
	}()
}

func main() {
	// set how many processes (threads to use)
	runtime.GOMAXPROCS(NCPU)

	app := cli.NewApp()
	app.Name = "stockscraper"
	app.Usage = "Scrap stock prices from Google."
	app.Flags = []cli.Flag{
		&cli.StringSliceFlag{
			Name: "name",
			Aliases: []string{"n"},
			Value: cli.NewStringSlice("reliance"),
			Usage: "Name of stock(s).",
		},
		&cli.BoolFlag{
			Name: "std",
			Value: false,
			Usage: "Use Personal Stock List for scraping Google.",
		},
	}
	app.Action = func(c *cli.Context) error {
		var wg sync.WaitGroup
		// create a new instance of the crawler structure
		crawler := &Crawler{
		}

		if c.Bool("std") {
			crawler.start(&wg, personalList...)
		} else {
			crawler.start(&wg, c.StringSlice("name")...)
		}
		// crawler.monitor(&wg, "reliance")
	
		wg.Wait()
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Received error while trying to run stockscraper: %v", err)
	}

	fmt.Println()
}
