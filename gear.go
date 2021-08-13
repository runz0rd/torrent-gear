package gear

import (
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
	Type  string
	Value string
}

func NewGerResult(value string) GearResult {
	type_ := GearResultTypeContent
	if strings.HasPrefix(value, "http") {
		type_ = GearResultTypeUrl
	}
	return GearResult{type_, value}
}

type GearType interface {
	// source could be an url or a filepath, depending on the implementation
	Process(source string) (results []GearResult, err error)
}

type GearConfig struct {
	Name            string `yaml:"name,omitempty"`
	Url             string `yaml:"url,omitempty"`
	CheckSecond     int    `yaml:"check_second,omitempty"`
	DestionationDir string `yaml:"destination_dir,omitempty"`
	Type            string `yaml:"type,omitempty"`
}

func (gc GearConfig) Handler() (GearType, error) {
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
		go func(conf GearConfig) {
			for {
				if err := g.handle(conf); err != nil {
					g.errHandler(err)
				}
			}
		}(gc)
	}
}

func (g *Gear) handle(gc GearConfig) error {
	log.Printf("[%v] checking", gc.Name)
	handler, err := gc.Handler()
	if err != nil {
		return err
	}
	results, err := handler.Process(gc.Url)
	if err != nil {
		return err
	}
	for _, result := range results {
		switch result.Type {
		case GearResultTypeUrl:
			if err := g.tc.AddFromUrl(result.Value, gc.DestionationDir); err != nil {
				return err
			}
		case GearResultTypeContent:
			if err := g.tc.AddContent([]byte(result.Value), gc.DestionationDir); err != nil {
				return err
			}
		default:
			return errors.Errorf("unsupported results type %q", result.Type)
		}
		log.Printf("[%v] added %q to torrent client", gc.Name, result)
	}
	time.Sleep(time.Duration(gc.CheckSecond) * time.Second)
	return nil
}
