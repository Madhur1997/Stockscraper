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

// how many threads to use within the application
const NCPU = 1

var _ sync.WaitGroup
var personalList []string = []string{"reliance", "ashok leyland", "indigo", "kesoram", "hdfc bank", "adani green energy",
	"vodafone idea", "TCS", "divis labs",}

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
