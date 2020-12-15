package main

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

func (crawler *Crawler) ppCmd() {
	for s, _ := range crawler.stocks {
		log.WithFields(log.Fields{
			"Command": strings.Title(crawler.ctx.Command.FullName()),
			"Stock": strings.ReplaceAll(strings.Title(s), "\"", ""),
		}).Info()
	}
}
