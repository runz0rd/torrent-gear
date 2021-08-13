package gear

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
)

type FeedGear struct {
	p *gofeed.Parser
}

func NewFeedGear() *FeedGear {
	parser := gofeed.NewParser()
	return &FeedGear{p: parser}
}

func (fg *FeedGear) Process(url string) ([]string, error) {
	var paths []string
	feed, err := fg.p.ParseURL(url)
	if err != nil {
		return nil, errors.Wrap(err, "parse error")
	}
	for _, item := range feed.Items {
		path, err := download(item.Link, os.TempDir())
		if err != nil {
			return nil, errors.Wrap(err, "download error")
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func download(fileUrl string, destDir string) (string, error) {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(fileUrl)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer resp.Body.Close()

	// Build fileName from fileUrl
	url, err := url.Parse(resp.Request.URL.Path)
	if err != nil {
		return "", errors.WithStack(err)
	}
	segments := strings.Split(url.Path, "/")
	fileName := segments[len(segments)-1]

	// Create file
	file, err := os.Create(path.Join(destDir, fileName))
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return path.Join(destDir, fileName), nil
}
