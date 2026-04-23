package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/sethvargo/go-githubactions"

	"github.com/leonhfr/discussions-action/action"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	gha := githubactions.New()

	if err := run(ctx, gha); err != nil {
		gha.Fatalf("error: %s", err)
	}
}

const (
	domainNameInput = "domain_name"
	targetDirInput  = "target_dir"
)

func run(ctx context.Context, gha *githubactions.Action) error {
	ghc, err := gha.Context()
	if err != nil {
		return err
	}

	domainName := gha.GetInput(domainNameInput)
	if domainName == "" {
		return fmt.Errorf("%s required", domainNameInput)
	}

	targetDir := gha.GetInput(targetDirInput)
	targetDir = filepath.Join(ghc.Workspace, targetDir)

	return action.Generate(ctx, gha, domainName, targetDir)
}
