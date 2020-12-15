package main

import (
	"github.com/urfave/cli/v2"
)

var debugFlag *cli.BoolFlag
var thresholdFlag *cli.IntFlag
var cmdFlags []cli.Flag

func init() {
	debugFlag = &cli.BoolFlag{
		Name: "debug",
		Value: false,
		Aliases: []string{"d"},
		Usage: "Print debug log(s)",
	}
	thresholdFlag = &cli.IntFlag{
				Name: "threshold",
				Aliases: []string{"t"},
				Value: 5,
				Usage: "Alert trigger threshold.",
			}
	cmdFlags = []cli.Flag{
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
}
