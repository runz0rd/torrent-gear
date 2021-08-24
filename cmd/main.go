package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/pkg/errors"
	gear "github.com/runz0rd/torrent-gear"
)

func main() {
	var flagConfig string
	var flagDebug bool
	flag.StringVar(&flagConfig, "config", "config.yaml", "config file location")
	flag.BoolVar(&flagDebug, "debug", false, "debug mode")
	flag.Parse()

	if err := run(flagConfig, flagDebug); err != nil {
		log.Fatal(err)
	}
}

func run(config string, debug bool) error {
	c, err := gear.ReadConfig(config)
	if err != nil {
		return errors.WithMessage(err, "config read error")
	}
	tc, err := c.Client.NewTorrentClientByType()
	if err != nil {
		return errors.WithMessage(err, "torrent client init error")
	}
	g := gear.NewGear(tc, func(err error) {
		log.Println(err)
		if debug {
			var stackTrace []string
			if err, ok := err.(gear.StackTracer); ok {
				for _, f := range err.StackTrace() {
					stackTrace = append(stackTrace, fmt.Sprintf("%+s:%d", f, f))
				}
			}

			if len(stackTrace) > 0 {
				log.Printf("stack: %v", strings.Join(stackTrace, "\n"))
			}
		}
	})
	g.Shift(c.Gears...)
	<-make(chan int) // block main thread
	return nil
}
