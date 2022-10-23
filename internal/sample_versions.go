package internal

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/mod/semver"
)

var (
	curMinVersion = mustCurMinVersion()

	//go:embed sample-stub-header
	sampleStubHeader string
)

func BuildSampleVersionsCommand() *cli.Command {
	return &cli.Command{
		Name: "sample-versions",
		Flags: []cli.Flag{
			&cli.PathFlag{Name: "from"},
			&cli.StringFlag{Name: "min", Value: curMinVersion},
		},
		Action: func(cCtx *cli.Context) error {
			return generateSampleVersions(
				cCtx.Context,
				cCtx.App.Writer,
				cCtx.String("min"),
				cCtx.Path("from"),
			)
		},
	}
}

func generateSampleVersions(ctx context.Context, out io.Writer, minBin, fromPath string) error {
	knownVersions, err := readCommentFiltered(fromPath)
	if err != nil {
		return err
	}

	minBin = withV(minBin)
	keepersByMajorMinor := map[string][]string{}

	for _, v := range knownVersions {
		v = withV(v)

		if !semver.IsValid(v) ||
			semver.Prerelease(v) != "" ||
			semver.Build(v) != "" ||
			semver.Compare(semver.Canonical(v), minBin) < 1 {
			continue
		}

		mmv := semver.MajorMinor(v)

		if _, ok := keepersByMajorMinor[mmv]; !ok {
			keepersByMajorMinor[mmv] = []string{}
		}

		keepersByMajorMinor[mmv] = append(keepersByMajorMinor[mmv], v)
	}

	keepersMap := map[string]struct{}{}

	for mmv := range keepersByMajorMinor {
		kbmmv := keepersByMajorMinor[mmv]

		sort.Sort(sort.Reverse(semver.ByVersion(kbmmv)))

		keepersMap[withoutV(kbmmv[0])] = struct{}{}
	}

	return writeLines(out, append([]string{sampleStubHeader}, revSortMapToSlice(keepersMap)...))
}

func mustCurMinVersion() string {
	curVersion := strings.TrimPrefix(runtime.Version(), "go")
	curMajor := semver.Major(withV(curVersion))
	curMinor := strings.TrimPrefix(semver.MajorMinor(withV(curVersion)), curMajor+".")

	curMinorInt, err := strconv.ParseInt(curMinor, 10, 64)
	if err != nil {
		panic(err)
	}

	minMinorInt := curMinorInt - 4
	if minMinorInt < 0 {
		minMinorInt = 0
	}

	return withoutV(fmt.Sprintf("%[1]v.%[2]v", curMajor, minMinorInt))
}

func withV(s string) string {
	return "v" + withoutV(s)
}

func withoutV(s string) string {
	return strings.TrimPrefix(s, "v")
}
