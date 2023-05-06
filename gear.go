package gear

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/runz0rd/common-go"
	"gopkg.in/yaml.v3"
)

type StackTracer interface {
	StackTrace() errors.StackTrace
}

const (
	GearResultTypeContent = "content"
	GearResultTypeUrl     = "url"
)

type GearResult struct {
	Name  string
	Type  string
	Value string
}

func NewGearResult(name, value string) GearResult {
	type_ := GearResultTypeContent
	if strings.HasPrefix(value, "http") {
		type_ = GearResultTypeUrl
	}
	return GearResult{name, type_, value}
}

type GearHandler interface {
	// source could be an url or a filepath, depending on the implementation
	Handle(source string) (results []GearResult, err error)
}

type GearConfig struct {
	Name            string `yaml:"name,omitempty"`
	Url             string `yaml:"url,omitempty"`
	CheckSecond     int    `yaml:"check_second,omitempty"`
	DestionationDir string `yaml:"destination_dir,omitempty"`
	Type            string `yaml:"type,omitempty"`
}

func (gc GearConfig) GearHandler() (GearHandler, error) {
	switch gc.Type {
	// todo implement more gears!
	case "feed":
		return NewFeedGear(), nil
	}
	return nil, errors.Errorf("no handler of type %v found", gc.Type)
}

type Config struct {
	Client common.TorrentClientConfig `yaml:"client,omitempty"`
	Gears  []GearConfig               `yaml:"gears,omitempty"`
}

func ReadConfig(path string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

type Gear struct {
	tc         common.TorrentClient
	errHandler func(err error)
}

func NewGear(tc common.TorrentClient, errHandler func(err error)) *Gear {
	return &Gear{tc, errHandler}
}

func (g *Gear) Shift(gcs ...GearConfig) {
	for _, gc := range gcs {
		go g.handle(gc)
	}
}

func (g *Gear) handle(gc GearConfig) {
	for {
		log.Println(gc.wrapMessagef("checking"))
		handler, err := gc.GearHandler()
		if err != nil {
			g.errHandler(errors.Wrap(err, gc.wrapMessagef("handler init error")))
			return
		}
		results, err := handler.Handle(gc.Url)
		if err != nil {
			g.errHandler(errors.Wrap(err, gc.wrapMessagef("handler error")))
			return
		}
		for _, result := range results {
			switch result.Type {
			case GearResultTypeUrl:
				if err := g.tc.AddFromUrl(result.Value, gc.DestionationDir); err != nil {
					gearErr := gc.wrapMessagef("torrent client error for %q of type %q", result.Name, result.Type)
					g.errHandler(errors.Wrap(err, gearErr))
					continue
				}
			case GearResultTypeContent:
				if err := g.tc.AddContent([]byte(result.Value), gc.DestionationDir); err != nil {
					gearErr := gc.wrapMessagef("torrent client error for %q of type %q", result.Name, result.Type)
					g.errHandler(errors.Wrap(err, gearErr))
					continue
				}
			default:
				g.errHandler(errors.Errorf(gc.wrapMessagef("unsupported result type %q for %q", result.Type, result.Name)))
				continue
			}
			log.Println(gc.wrapMessagef("added %q to torrent client", result.Name))
		}
		time.Sleep(time.Duration(gc.CheckSecond) * time.Second)
	}
}

func (gc GearConfig) wrapMessagef(format string, args ...interface{}) string {
	return fmt.Sprintf("[%v] %v", gc.Name, fmt.Sprintf(format, args...))
}
