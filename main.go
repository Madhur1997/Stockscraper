package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/urfave/cli/v2"
)

// How many threads to use within the application
const NCPU = 1

var _ sync.WaitGroup
var personalList []string = []string{"reliance", "ashok leyland", "indigo", "kesoram", "hdfc bank", "adani green energy",
	"vodafone idea", "divis labs",}

func main() {
	// set how many processes (threads to use)
	runtime.GOMAXPROCS(NCPU)

	done := make(chan bool)
	app := cli.NewApp()
	app.Name = "stockscraper"
	app.Usage = "Scrap stock prices from Google."
	cmnFlags := []cli.Flag{
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
	thresholdFlag := &cli.IntFlag{
				Name: "threshold",
				Aliases: []string{"t"},
				Value: 5,
				Usage: "Alert trigger threshold.",
			}

	app.Commands = []*cli.Command{
		&cli.Command{
			Name: "scrap",
			Usage: "Scrap some stock(s)",
			Action: func(c *cli.Context) error {
					var crawler *Crawler
					if c.Bool("std") {
						crawler = NewCrawler(c, personalList...)
					} else {
						crawler = NewCrawler(c, c.StringSlice("name")...)
					}

					go crawler.scrapStockPrices(done)
					select {
						case <-done: log.Println("Exiting")
					}
				
					return nil
				},
			Flags: cmnFlags,
		},
		&cli.Command{
			Name: "monitor",
			Usage: "Monitor some stock(s)",
			Action: func(c *cli.Context) error {

					var crawler *Crawler
					exit := make(chan os.Signal)
					signal.Notify(exit, syscall.SIGINT)

					if c.Bool("std") {
						crawler = NewCrawler(c, personalList...)
					} else {
						crawler = NewCrawler(c, c.StringSlice("name")...)
					}

					go crawler.monitor(done, exit)
					select {
						case <-done: log.Println("Exiting")
					}
				
					return nil
				},
			Flags: append(cmnFlags, thresholdFlag),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Received error while trying to run stockscraper: %v", err)
	}

	fmt.Println()
}
