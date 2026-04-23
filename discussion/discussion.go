package discussion

import "context"

type Fetcher interface {
	String() string
	Fetch(ctx context.Context, domainName string, existing []Discussion) ([]Discussion, error)
}

type Discussion struct {
	Title        string  `toml:"title"`
	URL          string  `toml:"url"`
	BlogRelURL   string  `toml:"blog_rel_url"`
	Forum        string  `toml:"forum"`
	SubForum     *string `toml:"sub_forum,omitempty"`
	Timestamp    int     `toml:"timestamp"`
	CommentCount int     `toml:"comment_count"`
	Score        int     `toml:"score"`
}
