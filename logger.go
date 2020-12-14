package main

import (
	"fmt"
	"log"
	"strings"
)

func (crawler *Crawler) ppCmd() {
	for s, _ := range crawler.stocks {
		log.Printf("%s %s\n", strings.Title(crawler.ctx.Command.FullName()), strings.Title(s))
	}
	fmt.Println()
}
