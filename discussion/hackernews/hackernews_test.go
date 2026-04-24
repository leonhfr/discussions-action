package hackernews

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/leonhfr/discussions-action/discussion"
)

func TestFetch(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/response.json")
	if err != nil {
		t.Fatal(err)
	}

	linter := discussion.Discussion{
		Title:        "I shipped a transaction bug, so I built a linter",
		URL:          "https://news.ycombinator.com/item?id=47715389",
		BlogRelURL:   "/posts/go-transaction-linter/",
		Forum:        "hackernews",
		Timestamp:    1775811665,
		CommentCount: 12,
		Score:        53,
	}

	older := discussion.Discussion{
		Title: "An older post",
		URL:   "https://leonh.fr/posts/older/",
		Forum: "hackernews",
		Score: 10,
	}

	tests := []struct {
		name     string
		existing []discussion.Discussion
		want     []discussion.Discussion
	}{
		{
			name:     "no existing",
			existing: nil,
			want:     []discussion.Discussion{linter},
		},
		{
			name:     "existing preserved when not refetched",
			existing: []discussion.Discussion{older},
			want:     []discussion.Discussion{linter, older},
		},
		{
			name: "existing replaced when refetched",
			existing: []discussion.Discussion{
				{URL: "https://news.ycombinator.com/item?id=47715389", Forum: "hackernews", Score: 999},
			},
			want: []discussion.Discussion{linter},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(data)
			}))
			defer server.Close()

			hn := HackerNews{restyClient: resty.New(), baseURL: server.URL}

			got, err := hn.Fetch(context.Background(), "leonh.fr", tt.existing)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
