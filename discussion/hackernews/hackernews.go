package hackernews

import (
	"context"
	"net/url"
	"slices"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"

	"github.com/leonhfr/discussions-action/discussion"
)

const (
	baseURL = "https://hn.algolia.com/api/v1/search_by_date"
	site    = "hackernews"
)

type HackerNews struct {
	restyClient *resty.Client
	baseURL     string
}

func New(restyClient *resty.Client) HackerNews {
	return HackerNews{
		restyClient: restyClient,
		baseURL:     baseURL,
	}
}

func (HackerNews) String() string { return site }

func (hn HackerNews) Fetch(ctx context.Context, domainName string, existing []discussion.Discussion) ([]discussion.Discussion, error) {
	var results []discussion.Discussion
	numberOfPages := 1
	for page := 0; page < numberOfPages; page++ {
		var resp response
		_, err := hn.restyClient.R().
			SetContext(ctx).
			SetResult(&resp).
			SetQueryParams(map[string]string{
				"query":                        domainName,
				"tags":                         "story",
				"page":                         strconv.Itoa(page),
				"restrictSearchableAttributes": "url",
			}).
			Get(hn.baseURL)
		if err != nil {
			return nil, err
		}

		results = append(results, lo.Map(resp.Hits, hitToDiscussion)...)
		numberOfPages = resp.NumberOfPages
	}

	return mergeWithExisting(results, existing), nil
}

func mergeWithExisting(fetched, existing []discussion.Discussion) []discussion.Discussion {
	fetchedMap := lo.SliceToMap(fetched, func(d discussion.Discussion) (string, bool) { return d.URL, true })
	existingFiltered := lo.Filter(existing, func(d discussion.Discussion, _ int) bool { return !fetchedMap[d.URL] })
	return slices.Concat(fetched, existingFiltered)
}

type response struct {
	Hits          []hit `json:"hits"`
	NumberOfPages int   `json:"nbPages"`
	Page          int   `json:"page"`
}

type hit struct {
	Title            string `json:"title"`
	URL              string `json:"url"`
	Points           int    `json:"points"`
	NumberOfComments int    `json:"num_comments"`
	CreatedAt        int    `json:"created_at_i"`
}

func hitToDiscussion(h hit, _ int) discussion.Discussion {
	blogRelURL := ""
	if u, err := url.Parse(h.URL); err == nil {
		blogRelURL = u.Path
	}
	return discussion.Discussion{
		Title:        h.Title,
		URL:          h.URL,
		BlogRelURL:   blogRelURL,
		Forum:        site,
		Timestamp:    h.CreatedAt,
		CommentCount: h.NumberOfComments,
		Score:        h.Points,
	}
}
