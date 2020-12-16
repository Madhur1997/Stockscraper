package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/urfave/cli/v2"
)

// How many threads to use within the application
const NCPU = 8
var _ sync.WaitGroup

func main() {
	// set how many processes (threads to use)
	runtime.GOMAXPROCS(NCPU)

	done := make(chan bool)
	app := cli.NewApp()
	app.Name = "stockscraper"
	app.Usage = "Scrap stock prices from Google."
	app.Flags = []cli.Flag{
			
		}
	app.Action = func(c *cli.Context) error {
			setLogger(c)
			return nil
		}

	app.Commands = []*cli.Command{
		&cli.Command{
			Name: "scrap",
			Usage: "Scrap some stock(s)",
			Action: func(c *cli.Context) error {
					setLogger(c)
					var crawler *Crawler
					if c.Bool("std") {
						crawler = NewCrawler(c, personalList...)
					} else {
						crawler = NewCrawler(c, c.StringSlice("name")...)
					}

					go crawler.scrapStockPrices(done)
					select {
						case <-done: log.Warn("Exiting")
					}
				
					return nil
				},
			Flags: append(cmdFlags, debugFlag),
		},
		&cli.Command{
			Name: "monitor",
			Usage: "Monitor some stock(s)",
			Action: func(c *cli.Context) error {
					var crawler *Crawler
					setLogger(c)
					exit := make(chan os.Signal)
					signal.Notify(exit, syscall.SIGINT)

					if c.Bool("std") {
						crawler = NewCrawler(c, personalList...)
					} else {
						crawler = NewCrawler(c, c.StringSlice("name")...)
					}

					go crawler.monitor(done, exit)
					select {
						case <-done: log.Warn("Exiting")
					}
				
					return nil
				},
			Flags: append(cmdFlags, thresholdFlag, debugFlag),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Received error while trying to run stockscraper: %v", err)
	}

	fmt.Println()
}
