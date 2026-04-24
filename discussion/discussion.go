package discussion

import (
	"context"

	"github.com/samber/lo"
)

type Fetcher interface {
	String() string
	Fetch(ctx context.Context, domainName string, existing []Discussion) ([]Discussion, error)
}

func FilterByForum(discussions []Discussion, forum string) []Discussion {
	return lo.Filter(discussions, func(d Discussion, _ int) bool { return d.Forum == forum })
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
