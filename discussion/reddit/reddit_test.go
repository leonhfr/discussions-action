package reddit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/leonhfr/discussions-action/discussion"
)

type noopLogger struct{}

func (noopLogger) Errorf(string, ...any) {}

func newServer(t *testing.T, path string) *httptest.Server {
	t.Helper()
	path = filepath.Clean(path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
}

func TestFetch(t *testing.T) {
	t.Parallel()

	sub := "r/golang"
	linter := discussion.Discussion{
		Title:        "I shipped a transaction bug, so I built a linter",
		URL:          "https://www.reddit.com/r/golang/comments/1qp93bs/i_shipped_a_transaction_bug_so_i_built_a_linter/",
		BlogRelURL:   "/posts/go-transaction-linter/",
		Forum:        "reddit",
		SubForum:     &sub,
		Timestamp:    1769600519,
		CommentCount: 12,
		Score:        53,
	}

	linterHighScore := linter
	linterHighScore.Score = 100

	older := discussion.Discussion{
		URL:   "https://www.reddit.com/r/golang/comments/older/",
		Forum: "reddit",
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
			name: "existing score preserved when higher",
			existing: []discussion.Discussion{
				{URL: linter.URL, Forum: "reddit", Score: 100},
			},
			want: []discussion.Discussion{linterHighScore},
		},
		{
			name: "fetched score used when higher than existing",
			existing: []discussion.Discussion{
				{URL: linter.URL, Forum: "reddit", Score: 1},
			},
			want: []discussion.Discussion{linter},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := newServer(t, "testdata/response.json")
			defer server.Close()

			r := Reddit{restyClient: resty.New(), baseURL: server.URL, apifyToken: "", logger: noopLogger{}}

			got, err := r.Fetch(context.Background(), "leonh.fr", tt.existing)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
