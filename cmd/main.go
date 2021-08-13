package main

import (
	"flag"
	"log"

	"github.com/pkg/errors"
	gear "github.com/runz0rd/torrent-gear"
)

func main() {
	var flagConfig string
	flag.StringVar(&flagConfig, "config", "config.yaml", "config file location")
	flag.Parse()

	if err := run(flagConfig); err != nil {
		log.Fatal(err)
	}
}

func run(config string) error {
	c, err := gear.ReadConfig(config)
	if err != nil {
		return errors.Wrap(err, "config read error")
	}
	tc, err := c.Client.NewTorrentClientByType()
	if err != nil {
		return errors.Wrap(err, "torrent client init error")
	}
	g := gear.NewGear(tc, func(err error) {
		if err, ok := err.(gear.StackTracer); ok {
			for _, f := range err.StackTrace() {
				log.Printf("%+s:%d\n", f, f)
			}
		} else {
			log.Print(err)
		}
	})
	g.Shift(c.Gears...)
	<-make(chan int) // block main thread
	return nil
}
