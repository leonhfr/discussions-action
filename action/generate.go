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

type Config struct {
	DomainName string
	TargetDir  string
	ApifyToken string
}

type Logger interface {
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
	Infof(format string, args ...any)
}

func Generate(ctx context.Context, config Config, logger Logger) error {
	limiter := rate.NewLimiter(rate.Every(time.Second), 10)

	restyClient := resty.
		New().
		OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
			return limiter.Wait(ctx)
		})

	fetchers := []discussion.Fetcher{
		hackernews.New(restyClient),
		reddit.New(restyClient, logger, config.ApifyToken),
	}

	target := filepath.Join(config.TargetDir, "discussions.toml")

	existingDiscussions, err := readOutput(target)
	if err != nil {
		return err
	}
	logger.Infof("loaded %d existing discussions from %s", len(existingDiscussions), target)

	var fetchedDiscussions []discussion.Discussion
	for _, fetcher := range fetchers {
		logger.Infof("fetching discussions from %s", fetcher.String())
		forum := fetcher.String()
		existing := discussion.FilterByForum(existingDiscussions, forum)
		fetched, err := fetcher.Fetch(ctx, config.DomainName, existing)
		if err != nil {
			logger.Errorf("failed to fetch from %s: %s", forum, err)
			return err
		}
		logger.Infof("fetched %d discussions from %s", len(fetched), forum)

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
	Discussions []discussion.Discussion `toml:"discussions"`
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
