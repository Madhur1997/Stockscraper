package main

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func setLogLevel(c *cli.Context) {
	if c.Bool("d") {
		log.SetLevel(log.DebugLevel)
	}
}

func (crawler *Crawler) ppCmd() {
	for s, _ := range crawler.stocks {
		log.WithFields(log.Fields{
			"Command": strings.Title(crawler.ctx.Command.FullName()),
			"Stock": strings.ReplaceAll(strings.Title(s), "\"", ""),
		}).Info()
	}
}
