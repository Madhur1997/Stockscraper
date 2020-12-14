package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/urfave/cli/v2"
)

// Our crawler structure definition
type Crawler struct {
	ctx *cli.Context
	alertThreshold int
	stocks map[string][]float64
	
	sync.Mutex
}

func NewCrawler(c *cli.Context, stocks ...string) *Crawler {
	crawler := &Crawler{ctx: c,}
	crawler.stocks = make(map[string][]float64)
	for _, s := range stocks {
		crawler.stocks[s] = make([]float64, 10000)
	}
	
	threshold := c.Int("threshold")
	if threshold != 0 {
		crawler.alertThreshold = threshold
	}
		
	return crawler
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

func (crawler *Crawler) spawnScrapers(res chan<- string) {
	crawler.Lock()
	defer crawler.Unlock()

	for s, _ := range crawler.stocks {
		go fetchPriceFromGoogle(s, res)
	}
}

func (crawler *Crawler) scrapStockPrices(done chan<- bool) {

	crawler.ppCmd()

	res := make(chan string)
	crawler.spawnScrapers(res)

	crawler.Lock()
	defer crawler.Unlock()
	for i := 0; i < len(crawler.stocks); i++ {
		select {
			case val := <-res:
				if val != "" {
					log.Println(val)
				}
		}
	}
	done <- true
}

func (crawler *Crawler) analyze(val string) {

	crawler.Lock()
	defer crawler.Unlock()

	valSlice := strings.Split(val, ":")
	stock := valSlice[0]
	temp := strings.Replace(strings.TrimSpace(valSlice[1]), ",", "", -1)

	stockPrice, err := strconv.ParseFloat(temp, 32)
	if err != nil {
		log.Fatalf("Error while converting string %s to float: %v", temp, err)
	}

	crawler.stocks[stock] = append(crawler.stocks[stock], stockPrice)

	incCt := 0
	decCt := 0
	for idx := len(crawler.stocks[stock])-1; idx > 0; idx-- {
		if crawler.stocks[stock][idx] > crawler.stocks[stock][idx - 1] {
			incCt++;
		}
	}

	for idx := len(crawler.stocks[stock])-2; idx >= 0; idx-- {
		if crawler.stocks[stock][idx] < crawler.stocks[stock][idx + 1] {
			decCt++;
		}
	}

	if incCt >= crawler.alertThreshold {
		log.Printf("%s has made consistent upward movements in last %d captures", stock, crawler.alertThreshold)
	}

	if decCt >= crawler.alertThreshold {
		log.Printf("%s has made consistent downward movements in last %d captures", stock, crawler.alertThreshold)
	}
}

func (crawler *Crawler) monitor(done chan<- bool, exit <-chan os.Signal) {

	crawler.ppCmd()
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
			case <-ticker.C:
				res := make(chan string)
				crawler.spawnScrapers(res)
				crawler.Lock()
				length := len(crawler.stocks)
				crawler.Unlock()
				for i := 0; i <length; i++ {
					select {
						case val := <-res:
							if val != "" {
								log.Println(val)
								go crawler.analyze(val)
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
