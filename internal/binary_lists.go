package internal

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/mod/semver"
	"golang.org/x/net/html"
)

var (
	curMinVersion = mustCurMinVersion()
	goLinkBaseURL = &url.URL{Scheme: "https", Host: "go.dev", Path: "/dl/"}

	//go:embed sample-stub-header
	sampleStubHeader string

	//go:embed all-stub-header
	allStubHeader string
)

func BuildSampleBinaryListCommand() *cli.Command {
	return &cli.Command{
		Name: "sample-binary-list",
		Flags: []cli.Flag{
			&cli.PathFlag{Name: "from"},
			&cli.StringFlag{Name: "min", Value: curMinVersion},
		},
		Action: func(cCtx *cli.Context) error {
			return generateSampleBinaryList(
				cCtx.Context,
				cCtx.App.Writer,
				cCtx.String("min"),
				cCtx.Path("from"),
			)
		},
	}
}

func generateSampleBinaryList(ctx context.Context, out io.Writer, minBin, fromPath string) error {
	binVersions, err := readCommentFiltered(fromPath)
	if err != nil {
		return err
	}

	minBin = withV(minBin)
	keepersByMajorMinor := map[string][]string{}

	for _, v := range binVersions {
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
		bv := semver.ByVersion(keepersByMajorMinor[mmv])

		sort.Strings(bv)

		keepersMap[withoutV(bv[len(bv)-1])] = struct{}{}
	}

	return writeLines(out, append([]string{sampleStubHeader}, revSortMapToSlice(keepersMap)...))
}

func BuildBinaryListCommand() *cli.Command {
	return &cli.Command{
		Name: "binary-list",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "os"},
			&cli.PathFlag{Name: "from"},
		},
		Action: func(cCtx *cli.Context) error {
			return generateBinaryList(cCtx.Context, cCtx.App.Writer, cCtx.String("os"), cCtx.Path("from"))
		},
	}
}

func generateBinaryList(ctx context.Context, out io.Writer, osName, fromPath string) error {
	goLinksBytes, err := os.ReadFile(fromPath)
	if err != nil {
		return err
	}

	goBinsMap := map[string]struct{}{}

	for _, line := range strings.Split(string(goLinksBytes), "\n") {
		if !strings.Contains(line, osName) {
			continue
		}

		if !strings.HasSuffix(line, "tar.gz") {
			continue
		}

		goBin, err := binNameFromLink(line, osName)
		if err != nil {
			return err
		}

		goBinsMap[goBin] = struct{}{}
	}

	return writeLines(out, append([]string{allStubHeader}, revSortMapToSlice(goBinsMap)...))
}

func binNameFromLink(link, osName string) (string, error) {
	parsed, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	baseName := filepath.Base(parsed.Path)
	cutIndex := strings.Index(baseName, "."+osName)

	return strings.TrimPrefix(baseName[:cutIndex], "go"), nil
}

func BuildGoLinksCommand() *cli.Command {
	return &cli.Command{
		Name: "go-links",
		Action: func(cCtx *cli.Context) error {
			return generateGoLinks(cCtx.Context, cCtx.App.Writer)
		},
	}
}

func generateGoLinks(ctx context.Context, w io.Writer) error {
	goLinks, err := fetchGoLinks(ctx)
	if err != nil {
		return err
	}

	for _, link := range goLinks {
		fmt.Println(link)
	}

	return nil
}

func fetchGoLinks(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, goLinkBaseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	goLinks := []string{}

	addGoLink := func(s string) {
		parsed, err := url.Parse(s)
		if err != nil {
			log.Printf("WARN: ignoring link %[1]q (err=%[2]v)", s, err)
			return
		}

		goLinks = append(goLinks, goLinkBaseURL.ResolveReference(parsed).String())
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			isDownloadLink := false
			href := ""

			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "download" {
					isDownloadLink = true
				}

				if attr.Key == "href" {
					href = attr.Val
				}
			}

			if isDownloadLink {
				addGoLink(href)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	sort.Sort(sort.Reverse(sort.StringSlice(goLinks)))

	return goLinks, nil
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
