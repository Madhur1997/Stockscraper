package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/urfave/cli/v2"
)

// how many threads to use within the application
const NCPU = 1

var _ sync.WaitGroup
var personalList []string = []string{"reliance", "ashok leyland", "indigo", "kesoram", "hdfc bank", "adani green energy",
	"vodafone idea", "TCS", "divis labs",}

// Our crawler structure definition
type Crawler struct {
	ctx *cli.Context
}

func fetchPriceFromGoogle(q string, res chan<- string) {

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
	var result string
	err := chromedp.Run(
		timeoutContext,
		chromedp.Navigate(url),
		chromedp.WaitVisible(inTextSel),
		chromedp.SendKeys(inTextSel, inQ),
		chromedp.Click(btnSel, chromedp.ByQuery),
		chromedp.WaitVisible(outTextSel),
		chromedp.Text(outTextSel, &result),
	)

	if err != nil {
		log.Printf("Error while scrapping stock price for %s: %v", strings.Title(q), err)
		chromedp.FromContext(ctx).Allocator.Wait()
		res <- ""
	}

	re := regexp.MustCompile("\\n")
	res <- re.ReplaceAllString(strings.Title(q) + ": " + result, " ")
}

func (crawler *Crawler) logCmd(queries ...string) {
	for _, q := range queries {
		log.Printf("%s %s\n", strings.Title(crawler.ctx.Command.FullName()), strings.Title(q))
	}
}

func (crawler *Crawler) spawnScrapers(res chan<- string, queries ...string) {

	for _, q := range queries {
		go fetchPriceFromGoogle(q, res)
	}
}

func (crawler *Crawler) scrapStockPrices(done chan<- bool, queries ...string) {

	crawler.logCmd(queries...)

	res := make(chan string)
	crawler.spawnScrapers(res, queries...)
	for i := 0; i < len(queries); i++ {
		select {
			case val := <-res:
				if val != "" {
					log.Println(val)
				}
		}
	}
	done <- true
}

func (crawler *Crawler) monitor(done chan<- bool, exit <-chan os.Signal, queries ...string) {

	crawler.logCmd(queries...)
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
			case <-ticker.C:
				res := make(chan string)
				crawler.spawnScrapers(res, queries...)
				for i := 0; i <len(queries); i++ {
					select {
						case val := <-res:
							if val != "" {
								log.Println(val)
							}
					}
				}
			case <-exit:
				log.Println("Received exit request, return.")
				done <- true
				return
		}
		fmt.Println()
	}
}

func main() {
	// set how many processes (threads to use)
	runtime.GOMAXPROCS(NCPU)

	done := make(chan bool)
	app := cli.NewApp()
	app.Name = "stockscraper"
	app.Usage = "Scrap stock prices from Google."
	flags := []cli.Flag{
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
	app.Commands = []*cli.Command{
		&cli.Command{
			Name: "scrap",
			Usage: "Scrap some stock(s)",
			Action: func(c *cli.Context) error {
					crawler := &Crawler{
						ctx: c,
					}

					if c.Bool("std") {
						go crawler.scrapStockPrices(done, personalList...)
					} else {
						go crawler.scrapStockPrices(done, c.StringSlice("name")...)
					}

					select {
						case <-done: log.Println("Exiting")
					}
				
					return nil
				},
			Flags: flags,
		},
		&cli.Command{
			Name: "monitor",
			Usage: "Monitor some stock(s)",
			Action: func(c *cli.Context) error {

					exit := make(chan os.Signal)
					signal.Notify(exit, syscall.SIGINT)
					crawler := &Crawler{
						ctx: c,
					}

					if c.Bool("std") {
						go crawler.monitor(done, exit, personalList...)
					} else {
						go crawler.monitor(done, exit, c.StringSlice("name")...)
					}

					select {
						case <-done: log.Println("Exiting")
					}
				
					return nil
				},
			Flags: flags,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Received error while trying to run stockscraper: %v", err)
	}

	fmt.Println()
}
