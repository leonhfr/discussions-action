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

	server := newServer(t, "testdata/response.json")
	defer server.Close()

	r := Reddit{restyClient: resty.New(), baseURL: server.URL}

	got, err := r.Fetch(context.Background(), "leonh.fr", nil)
	if err != nil {
		t.Fatal(err)
	}

	sub := "r/golang"
	want := []discussion.Discussion{
		{
			Title:        "I shipped a transaction bug, so I built a linter",
			URL:          "https://www.reddit.com/r/golang/comments/1qp93bs/i_shipped_a_transaction_bug_so_i_built_a_linter/",
			BlogRelURL:   "/posts/go-transaction-linter/",
			Forum:        "reddit",
			SubForum:     &sub,
			Timestamp:    1769600519,
			CommentCount: 12,
			Score:        54,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestFetch_existingScorePreservedWhenHigher(t *testing.T) {
	t.Parallel()

	server := newServer(t, "testdata/response.json")
	defer server.Close()

	r := Reddit{restyClient: resty.New(), baseURL: server.URL}

	sub := "r/golang"
	existing := []discussion.Discussion{
		{
			URL:   "https://www.reddit.com/r/golang/comments/1qp93bs/i_shipped_a_transaction_bug_so_i_built_a_linter/",
			Score: 100,
		},
	}

	got, err := r.Fetch(context.Background(), "leonh.fr", existing)
	if err != nil {
		t.Fatal(err)
	}

	want := []discussion.Discussion{
		{
			Title:        "I shipped a transaction bug, so I built a linter",
			URL:          "https://www.reddit.com/r/golang/comments/1qp93bs/i_shipped_a_transaction_bug_so_i_built_a_linter/",
			BlogRelURL:   "/posts/go-transaction-linter/",
			Forum:        "reddit",
			SubForum:     &sub,
			Timestamp:    1769600519,
			CommentCount: 12,
			Score:        100,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
