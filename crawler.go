package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/chromedp/chromedp"
	"github.com/faiface/beep"
	"github.com/urfave/cli/v2"
)

var _ beep.Buffer

// Our crawler structure definition
type Crawler struct {
	ctx *cli.Context
	alertThreshold int
	stocks map[string][]float64
	
	m sync.Mutex
}

func(crawler *Crawler) lock() {
	log.Debug("Aqcuire lock")
	crawler.m.Lock()
}

func(crawler *Crawler) unlock() {
	log.Debug("Release lock")
	crawler.m.Unlock()
}

func NewCrawler(c *cli.Context, stocks ...string) *Crawler {
	crawler := &Crawler{ctx: c,}
	crawler.stocks = make(map[string][]float64)
	for _, s := range stocks {
		crawler.stocks[strings.Title(s)] = make([]float64, 0)
	}
	
	threshold := c.Int("threshold")
	if threshold != 0 {
		crawler.alertThreshold = threshold
	}
		
	return crawler
}

func fetchPriceFromGoogle(q string, res chan<- string) {
	
	log.WithFields(log.Fields{
		"Stock": q,
	}).Debug("Fetch stock price")

	url := "https://www.google.com"
	inQ := q + " stock price"
	inTextSel := `//input[@name='q']`
	btnSel := `input[name="btnK"]`
	outTextSel := `//span[@jsname="vWLAgc"]`

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	// Wait for timeout.
	timeoutContext, _ := context.WithTimeout(ctx, 20 * time.Second)

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
		log.WithFields(log.Fields{
			"Stock": strings.Title(q),
			"Error": err,
		}).Error("Error scrapping stock price")
		res <- ""
		return
	}

	re := regexp.MustCompile("\\n")
	res <- re.ReplaceAllString(strings.Title(q) + ": " + result, " ")
}

func (crawler *Crawler) spawnScrapers(res chan<- string) {
	crawler.lock()
	defer crawler.unlock()

	for s, _ := range crawler.stocks {
		go fetchPriceFromGoogle(s, res)
	}
}

func (crawler *Crawler) scrapStockPrices(done chan<- bool) {

	crawler.ppCmd()

	res := make(chan string)
	crawler.spawnScrapers(res)

	crawler.lock()
	defer crawler.unlock()
	stkColl := make([]string, 0)
	for i := 0; i < len(crawler.stocks); i++ {
		select {
			case val := <-res:
				if val != "" {
					stkColl = append(stkColl, val)
				}
		}
	}
	sort.Strings(stkColl)
	for _, val := range stkColl {
		log.Info(val)
	}
	done <- true
}

func (crawler *Crawler) analyze(val string, wg *sync.WaitGroup) {
	log.WithFields(log.Fields{
		"Stock": val,
	}).Debug("Analyze stock")

	crawler.lock()
	defer crawler.unlock()
	defer wg.Done()

	valSlice := strings.Split(val, ":")
	stock := valSlice[0]
	temp := strings.Replace(strings.TrimSpace(valSlice[1]), ",", "", -1)

	stockPrice, err := strconv.ParseFloat(temp, 32)
	if err != nil {
		log.Fatalf("Error converting string %s to float: %v", temp, err)
	}

	crawler.stocks[stock] = append(crawler.stocks[stock], stockPrice)

	incCt := 0
	decCt := 0
	length := len(crawler.stocks[stock])
	for idx := length-1; idx > 0; idx-- {
		if crawler.stocks[stock][idx-1] >= crawler.stocks[stock][idx] {
			break;
		}
		incCt++;
	}

	for idx := length-2; idx >= 0; idx-- {
		if crawler.stocks[stock][idx] <= crawler.stocks[stock][idx + 1] {
			break
		}
		decCt++;
	}
	log.WithFields(log.Fields{
		"Stock": stock,
		"incCt": incCt,
		"decCt": decCt,
	}).Debug()

	incMaxInt := int(math.Max(float64(incCt), float64(crawler.alertThreshold)))
	decMaxInt := int(math.Max(float64(decCt), float64(crawler.alertThreshold)))
	if incCt >= crawler.alertThreshold {
		log.WithFields(log.Fields{
			"Stock": stock,
			"Interval": incMaxInt,
			"Prices": crawler.stocks[stock][length-incMaxInt:],
		}).Warn("Consistent upward movements\n")
	}

	if decCt >= crawler.alertThreshold {
		log.WithFields(log.Fields{
			"Stock": stock,
			"Interval": decMaxInt,
			"Prices": crawler.stocks[stock][length-decMaxInt:],
		}).Warn("Consistent downward movements\n")
	}
}

func (crawler *Crawler) monitor(done chan<- bool, exit <-chan os.Signal) {

	crawler.ppCmd()
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		select {
			case <-exit:
				log.Info("Exit request, return.")
				done <- true
				return
			}
	} ()
			
	for {
		select {
			case <-ticker.C:
				res := make(chan string)
				crawler.spawnScrapers(res)
				crawler.lock()
				length := len(crawler.stocks)
				crawler.unlock()
				stkColl := make([]string, 0)
				for i := 0; i <length; i++ {
					select {
						case val := <-res:
							if val != "" {
								stkColl = append(stkColl, val)
							}
					}
				}

				sort.Strings(stkColl)
				var wg sync.WaitGroup
				for _, val := range stkColl {
					log.Info(val)
					wg.Add(1)
					go crawler.analyze(val, &wg)
				}
				wg.Wait()
		}
		fmt.Println()
	}
}
