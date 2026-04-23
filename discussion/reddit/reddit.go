package reddit

import (
	"context"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"

	"github.com/leonhfr/discussions-action/discussion"
)

const (
	baseURL   = "https://www.reddit.com/search.json"
	redditURL = "https://www.reddit.com"
	site      = "reddit"
)

type Reddit struct {
	restyClient *resty.Client
	baseURL     string
}

func New(restyClient *resty.Client) Reddit {
	return Reddit{
		restyClient: restyClient,
		baseURL:     baseURL,
	}
}

func (Reddit) String() string { return site }

func (r Reddit) Fetch(ctx context.Context, domainName string, existing []discussion.Discussion) ([]discussion.Discussion, error) {
	existingByURL := lo.KeyBy(existing, func(d discussion.Discussion) string { return d.URL })

	var results []discussion.Discussion
	var after *string

	for {
		var resp response
		req := r.restyClient.R().
			SetContext(ctx).
			SetResult(&resp).
			SetQueryParams(map[string]string{
				"q":    "url:" + domainName,
				"type": "link",
				"sort": "new",
			})

		if after != nil {
			req.SetQueryParam("after", *after)
		}

		_, err := req.Get(r.baseURL)
		if err != nil {
			return nil, err
		}

		for _, child := range resp.Data.Children {
			d := child.Data.toDiscussion()
			if prev, ok := existingByURL[d.URL]; ok {
				d.Score = max(prev.Score, d.Score)
			}
			results = append(results, d)
		}

		if resp.Data.After == nil {
			break
		}
		after = resp.Data.After
	}

	return results, nil
}

type response struct {
	Data responseData `json:"data"`
}

type responseData struct {
	Children []child `json:"children"`
	After    *string `json:"after"`
}

type child struct {
	Data post `json:"data"`
}

type post struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Permalink   string  `json:"permalink"`
	Subreddit   string  `json:"subreddit_name_prefixed"`
	CreatedUTC  float64 `json:"created_utc"`
	NumComments int     `json:"num_comments"`
	Score       int     `json:"score"`
}

func (p post) toDiscussion() discussion.Discussion {
	blogRelURL := ""
	if u, err := url.Parse(p.URL); err == nil {
		blogRelURL = u.Path
	}
	subreddit := p.Subreddit
	return discussion.Discussion{
		Title:        p.Title,
		URL:          redditURL + p.Permalink,
		BlogRelURL:   blogRelURL,
		Forum:        site,
		SubForum:     &subreddit,
		Timestamp:    int(p.CreatedUTC),
		CommentCount: p.NumComments,
		Score:        p.Score,
	}
}
