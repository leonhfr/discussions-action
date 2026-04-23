package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/leonhfr/discussions-action/action"
)

const (
	domainName = "leonh.fr"
	targetDir  = "dist"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := action.Generate(ctx, fmtLogger{}, domainName, targetDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type fmtLogger struct{}

func (fmtLogger) Debugf(format string, args ...any) { fmt.Printf("DEBUG "+format+"\n", args...) }
func (fmtLogger) Errorf(format string, args ...any) { fmt.Printf("ERROR "+format+"\n", args...) }
func (fmtLogger) Infof(format string, args ...any)  { fmt.Printf("INFO  "+format+"\n", args...) }
