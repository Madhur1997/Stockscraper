package main

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func setLogger(c *cli.Context) {
	if c.Bool("d") {
		log.SetLevel(log.DebugLevel)
	}
	customFormatter := new(log.TextFormatter)
        customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)
}

func (crawler *Crawler) ppCmd() {
	for s, _ := range crawler.stocks {
		log.WithFields(log.Fields{
			"Command": strings.Title(crawler.ctx.Command.FullName()),
			"Stock": strings.ReplaceAll(strings.Title(s), "\"", ""),
		}).Info()
	}
}
