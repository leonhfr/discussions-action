package action

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"

	"github.com/leonhfr/discussions-action/discussion"
	"github.com/leonhfr/discussions-action/discussion/hackernews"
	"github.com/leonhfr/discussions-action/discussion/reddit"
)

type Logger interface {
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
	Infof(format string, args ...any)
}

func Generate(ctx context.Context, logger Logger, domainName, targetDir string) error {
	limiter := rate.NewLimiter(rate.Every(time.Second), 10)

	restyClient := resty.
		New().
		OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
			return limiter.Wait(ctx)
		})

	fetchers := []discussion.Fetcher{
		hackernews.New(restyClient),
		reddit.New(restyClient, logger),
	}

	target := filepath.Join(targetDir, "discussions.toml")

	existingDiscussions, err := readOutput(target)
	if err != nil {
		return err
	}
	logger.Infof("loaded %d existing discussions from %s", len(existingDiscussions), target)

	var fetchedDiscussions []discussion.Discussion
	for _, fetcher := range fetchers {
		logger.Infof("fetching discussions from %s", fetcher.String())
		fetched, err := fetcher.Fetch(ctx, domainName, existingDiscussions)
		if err != nil {
			logger.Errorf("failed to fetch from %s: %s", fetcher.String(), err)
			return err
		}
		logger.Infof("fetched %d discussions from %s", len(fetched), fetcher.String())

		fetchedDiscussions = slices.Concat(fetchedDiscussions, fetched)
	}

	slices.SortFunc(fetchedDiscussions, func(a, b discussion.Discussion) int {
		return b.Timestamp - a.Timestamp
	})
	sorted := fetchedDiscussions

	logger.Infof("writing %d discussions to %s", len(sorted), target)
	return writeOutput(target, sorted)
}

type output struct {
	Discussions []discussion.Discussion
}

func readOutput(path string) ([]discussion.Discussion, error) {
	path = filepath.Clean(path)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var output output
	if err := toml.Unmarshal(data, &output); err != nil {
		return nil, err
	}

	return output.Discussions, nil
}

func writeOutput(path string, discussions []discussion.Discussion) error {
	path = filepath.Clean(path)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	encoder := toml.NewEncoder(f)
	encoder.Indent = ""

	return encoder.Encode(output{Discussions: discussions})
}
