package action

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/leonhfr/discussions-action/discussion"
)

func TestReadOutput(t *testing.T) {
	t.Parallel()

	got, err := readOutput("testdata/discussions.toml")
	if err != nil {
		t.Fatal(err)
	}

	want := []discussion.Discussion{
		{
			Title:        "I shipped a transaction bug, so I built a linter",
			URL:          "https://news.ycombinator.com/item?id=47715389",
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

func TestReadOutput_missingFile(t *testing.T) {
	t.Parallel()

	got, err := readOutput("testdata/nonexistent.toml")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("got %+v, want nil", got)
	}
}

func TestWriteOutput(t *testing.T) {
	t.Parallel()

	discussions := []discussion.Discussion{
		{
			Title:        "I shipped a transaction bug, so I built a linter",
			URL:          "https://news.ycombinator.com/item?id=47715389",
			BlogRelURL:   "/posts/go-transaction-linter/",
			Forum:        "hackernews",
			Timestamp:    1775811665,
			CommentCount: 12,
			Score:        53,
		},
	}

	path := filepath.Join(t.TempDir(), "discussions.toml")

	if err := writeOutput(path, discussions); err != nil {
		t.Fatal(err)
	}

	got, err := readOutput(path)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, discussions) {
		t.Errorf("got %+v, want %+v", got, discussions)
	}
}

func TestWriteOutput_createsFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "discussions.toml")

	if err := writeOutput(path, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}
