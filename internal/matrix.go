package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v2"
)

type matrixEntry struct {
	Runner  string `json:"runner"`
	Target  string `json:"target"`
	Version string `json:"version"`
}

func BuildMatrixJSONCommand() *cli.Command {
	return &cli.Command{
		Name: "matrix-json",
		Flags: []cli.Flag{
			&cli.PathFlag{Name: "from", Value: "./.testdata/sample-versions.txt"},
		},
		Action: func(cCtx *cli.Context) error {
			return generateMatrixJSON(cCtx.Context, cCtx.Path("from"))
		},
	}
}

func generateMatrixJSON(ctx context.Context, sampleVersions string) error {
	matrixEntries := []matrixEntry{}

	for _, runner := range []string{"ubuntu-latest", "macos-latest"} {
		sampleVersionsSlice, err := readCommentFiltered(sampleVersions)
		if err != nil {
			return err
		}

		for _, target := range []string{"local"} { // FIXME: maybe get `arm` working?
			for _, v := range append([]string{"stable", "module", "master"}, sampleVersionsSlice...) {
				matrixEntries = append(
					matrixEntries,
					matrixEntry{
						Runner:  runner,
						Target:  target,
						Version: v,
					},
				)
			}
		}
	}

	var out io.Writer = os.Stdout
	asGithubOutput := false

	if v, ok := os.LookupEnv("GITHUB_OUTPUT"); ok {
		gho, err := os.Create(v)
		if err != nil {
			return err
		}

		defer gho.Close()

		asGithubOutput = true
		out = gho
	}

	if asGithubOutput {
		if _, err := fmt.Fprintf(out, "env<<EOF\n"); err != nil {
			return err
		}
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")

	if err := enc.Encode(matrixEntries); err != nil {
		return err
	}

	if asGithubOutput {
		if _, err := fmt.Fprintf(out, "EOF\n"); err != nil {
			return err
		}
	}

	return nil
}
