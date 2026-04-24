package reddit

import (
	"context"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"

	"github.com/leonhfr/discussions-action/discussion"
)

type Logger interface {
	Errorf(format string, args ...any)
}

const (
	baseURL = "https://api.apify.com/v2/acts/trudax~reddit-scraper-lite/run-sync-get-dataset-items"
	site    = "reddit"
)

type Reddit struct {
	restyClient *resty.Client
	baseURL     string
	apifyToken  string
	logger      Logger
}

func New(restyClient *resty.Client, logger Logger, apifyToken string) Reddit {
	return Reddit{
		restyClient: restyClient,
		baseURL:     baseURL,
		apifyToken:  apifyToken,
		logger:      logger,
	}
}

func (Reddit) String() string { return site }

func (r Reddit) Fetch(ctx context.Context, domainName string, existing []discussion.Discussion) ([]discussion.Discussion, error) {
	existingByURL := lo.KeyBy(existing, func(d discussion.Discussion) string { return d.URL })

	body := requestBody{
		DebugMode:           false,
		IgnoreStartUrls:     false,
		IncludeNSFW:         true,
		MaxComments:         10,
		MaxCommunitiesCount: 2,
		MaxItems:            100,
		MaxPostCount:        10,
		MaxUserCount:        2,
		Proxy: requestProxy{
			UseApifyProxy:    true,
			ApifyProxyGroups: []string{"RESIDENTIAL"},
		},
		ScrollTimeout:     40,
		SearchComments:    false,
		SearchCommunities: false,
		SearchPosts:       true,
		SearchUsers:       false,
		Searches:          []string{"url:" + domainName},
		SkipComments:      true,
		SkipCommunity:     false,
		SkipUserPosts:     true,
		Sort:              "new",
	}

	var posts []apifyPost
	httpResp, err := r.restyClient.R().
		SetContext(ctx).
		SetResult(&posts).
		SetBody(body).
		SetQueryParams(map[string]string{
			"token":             r.apifyToken,
			"memory":            "1024",
			"timeout":           "300",
			"maxTotalChargeUsd": "0.1",
		}).
		Post(r.baseURL)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsSuccess() {
		r.logger.Errorf("reddit: unexpected status: %d", httpResp.StatusCode())
	}

	var results []discussion.Discussion
	for _, p := range posts {
		d := p.toDiscussion()
		if prev, ok := existingByURL[d.URL]; ok {
			d.Score = max(prev.Score, d.Score)
		}
		results = append(results, d)
	}

	return results, nil
}

type requestBody struct {
	DebugMode           bool         `json:"debugMode"`
	IgnoreStartUrls     bool         `json:"ignoreStartUrls"`
	IncludeNSFW         bool         `json:"includeNSFW"`
	MaxComments         int          `json:"maxComments"`
	MaxCommunitiesCount int          `json:"maxCommunitiesCount"`
	MaxItems            int          `json:"maxItems"`
	MaxPostCount        int          `json:"maxPostCount"`
	MaxUserCount        int          `json:"maxUserCount"`
	Proxy               requestProxy `json:"proxy"`
	ScrollTimeout       int          `json:"scrollTimeout"`
	SearchComments      bool         `json:"searchComments"`
	SearchCommunities   bool         `json:"searchCommunities"`
	SearchPosts         bool         `json:"searchPosts"`
	SearchUsers         bool         `json:"searchUsers"`
	Searches            []string     `json:"searches"`
	SkipComments        bool         `json:"skipComments"`
	SkipCommunity       bool         `json:"skipCommunity"`
	SkipUserPosts       bool         `json:"skipUserPosts"`
	Sort                string       `json:"sort"`
}

type requestProxy struct {
	UseApifyProxy    bool     `json:"useApifyProxy"`
	ApifyProxyGroups []string `json:"apifyProxyGroups"`
}

type apifyPost struct {
	Title            string `json:"title"`
	URL              string `json:"url"`
	Link             string `json:"link"`
	CommunityName    string `json:"communityName"`
	CreatedAt        string `json:"createdAt"`
	NumberOfComments int    `json:"numberOfComments"`
	UpVotes          int    `json:"upVotes"`
}

func (p apifyPost) toDiscussion() discussion.Discussion {
	blogRelURL := ""
	if u, err := url.Parse(p.Link); err == nil {
		blogRelURL = u.Path
	}
	community := p.CommunityName
	timestamp := 0
	if t, err := time.Parse(time.RFC3339Nano, p.CreatedAt); err == nil {
		timestamp = int(t.Unix())
	}
	return discussion.Discussion{
		Title:        p.Title,
		URL:          p.URL,
		BlogRelURL:   blogRelURL,
		Forum:        site,
		SubForum:     &community,
		Timestamp:    timestamp,
		CommentCount: p.NumberOfComments,
		Score:        p.UpVotes,
	}
}
