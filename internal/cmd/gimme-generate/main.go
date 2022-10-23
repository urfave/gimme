package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
	"github.com/urfave/gimme/internal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app := &cli.App{
		Name:        "gimme-generate",
		Description: "internal generate-y tool for urfave/gimme",
		Commands: []*cli.Command{
			internal.BuildMatrixJSONCommand(),
			internal.BuildSampleVersionsCommand(),
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
