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
const NCPU = 8

var personalList []string = []string{"reliance", "ashok leyland", "indigo", "kesoram", "hdfc bank", "adani green energy",
	"vodafone idea", "TCS", "divis labs",}

// Our crawler structure definition
type Crawler struct {
	Ctx context.Context
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

	// Wait for timeout.
	timeoutContext, _ := context.WithTimeout(crawler.Ctx, 30 * time.Second)

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
		// create context
		ctx, _ := chromedp.NewContext(context.Background())
		// create a new instance of the crawler structure
		crawler := &Crawler{
			Ctx: ctx,
		}

		if c.Bool("std") {
			crawler.start(&wg, personalList...)
		} else {
			crawler.start(&wg, c.StringSlice("name")...)
		}
	
		wg.Wait()
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Received error while trying to run stockscraper: %v", err)
	}

	fmt.Println()
}
