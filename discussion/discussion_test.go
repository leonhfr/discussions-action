package discussion

import (
	"reflect"
	"testing"
)

func TestFilterByForum(t *testing.T) {
	t.Parallel()

	hn := Discussion{URL: "https://news.ycombinator.com/item?id=1", Forum: "hackernews"}
	rd := Discussion{URL: "https://www.reddit.com/r/golang/comments/abc/", Forum: "reddit"}

	tests := []struct {
		name        string
		discussions []Discussion
		forum       string
		want        []Discussion
	}{
		{
			name:        "returns only matching forum",
			discussions: []Discussion{hn, rd},
			forum:       "hackernews",
			want:        []Discussion{hn},
		},
		{
			name:        "returns empty when no match",
			discussions: []Discussion{rd},
			forum:       "hackernews",
			want:        []Discussion{},
		},
		{
			name:        "returns all when all match",
			discussions: []Discussion{hn, hn},
			forum:       "hackernews",
			want:        []Discussion{hn, hn},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FilterByForum(tt.discussions, tt.forum)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
