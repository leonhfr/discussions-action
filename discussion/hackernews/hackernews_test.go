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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	hn := HackerNews{restyClient: resty.New(), baseURL: server.URL}

	got, err := hn.Fetch(context.Background(), "leonh.fr", nil)
	if err != nil {
		t.Fatal(err)
	}

	want := []discussion.Discussion{
		{
			Title:        "I shipped a transaction bug, so I built a linter",
			URL:          "https://leonh.fr/posts/go-transaction-linter/",
			BlogRelURL:   "/posts/go-transaction-linter/",
			Forum:        "hackernews",
			Timestamp:    1775811665,
			CommentCount: 12,
			Score:        53,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
