package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
			&cli.PathFlag{Name: "top-dir", Value: "./"},
		},
		Action: func(cCtx *cli.Context) error {
			return generateMatrixJSON(cCtx.Context, cCtx.Path("top-dir"))
		},
	}
}

func generateMatrixJSON(ctx context.Context, topDir string) error {
	runnerGoos := map[string]string{
		"ubuntu-latest": "linux",
		"macos-latest":  "darwin",
	}

	matrixEntries := []matrixEntry{}

	for _, runner := range []string{"ubuntu-latest", "macos-latest"} {
		for _, target := range []string{"local"} { // FIXME: maybe get `arm` working?
			runnerVersions, err := readGoVersions(ctx, topDir, runnerGoos[runner])
			if err != nil {
				return err
			}

			for _, v := range runnerVersions {
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

func readGoVersions(ctx context.Context, topDir, goos string) ([]string, error) {
	versions := []string{}

	binVersions, err := readCommentFiltered(filepath.Join(topDir, ".testdata", "sample-binary-"+goos))
	if err != nil {
		return nil, err
	}

	return append(append(versions, binVersions...), "stable", "module", "master"), nil
}
