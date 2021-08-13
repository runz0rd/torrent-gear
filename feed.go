package gear

import (
	"github.com/mmcdole/gofeed"
)

type FeedGear struct {
	p *gofeed.Parser
}

func NewFeedGear() *FeedGear {
	parser := gofeed.NewParser()
	return &FeedGear{p: parser}
}

func (fg *FeedGear) Handle(url string) ([]GearResult, error) {
	var results []GearResult
	feed, err := fg.p.ParseURL(url)
	if err != nil {
		return nil, err
	}
	for _, item := range feed.Items {
		results = append(results, NewGerResult(item.Link))
	}
	return results, nil
}
