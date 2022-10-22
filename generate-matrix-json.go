package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

type matrixEntry struct {
	Runner  string `json:"runner"`
	Target  string `json:"target"`
	Version string `json:"version"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	topBytes, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		log.Fatal(err)
	}

	runnerGoos := map[string]string{
		"ubuntu-latest": "linux",
		"macos-latest":  "darwin",
	}

	top := strings.TrimSpace(string(topBytes))

	matrixEntries := []matrixEntry{}

	for _, runner := range []string{"ubuntu-latest", "macos-latest"} {
		for _, target := range []string{"local"} { // FIXME: maybe get `arm` working?
			runnerVersions, err := readGoVersions(ctx, top, runnerGoos[runner])
			if err != nil {
				log.Fatal(err)
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
			log.Fatal(err)
		}

		defer gho.Close()

		asGithubOutput = true
		out = gho
	}

	if asGithubOutput {
		if _, err := fmt.Fprintf(out, "env<<EOF\n"); err != nil {
			log.Fatal(err)
		}
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")

	if err := enc.Encode(matrixEntries); err != nil {
		log.Fatal(err)
	}

	if asGithubOutput {
		if _, err := fmt.Fprintf(out, "EOF\n"); err != nil {
			log.Fatal(err)
		}
	}
}

func readGoVersions(ctx context.Context, top, goos string) ([]string, error) {
	versions := []string{}

	binVersions, err := readCommentFiltered(filepath.Join(top, ".testdata", "sample-binary-"+goos))
	if err != nil {
		return nil, err
	}

	versions = append(versions, binVersions...)

	sourceVersions, err := readCommentFiltered(filepath.Join(top, ".testdata", "source-"+goos))
	if err != nil {
		return nil, err
	}

	return append(versions, sourceVersions...), nil
}

func readCommentFiltered(filename string) ([]string, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	keepers := []string{}

	for _, line := range strings.Split(string(fileBytes), "\n") {
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		keepers = append(keepers, line)
	}

	return keepers, nil
}
