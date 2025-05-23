package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

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
	// ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	// defer cancel()

	logger := log.Default()

	deps := web.Deps{
		Logger: logger,
		Addr:   ":3000",
	}

	if err := web.ListenAndServe(deps); err != nil {
		return err
	}

	return nil
}
