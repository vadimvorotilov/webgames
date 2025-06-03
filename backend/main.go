package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"webgames/internal/web"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	stdout io.Writer,
	args []string,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	if err := web.ListenAndServe(":3000"); err != nil {
		return err
	}

	return nil
}
