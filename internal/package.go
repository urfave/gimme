package internal

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

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

func writeLines(w io.Writer, lines []string) error {
	_, err := fmt.Fprintln(
		w,
		strings.Join(
			lines,
			"\n",
		),
	)

	return err
}

func revSortMapToSlice(sl map[string]struct{}) []string {
	uniq := []string{}

	for s := range sl {
		uniq = append(uniq, s)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(uniq)))

	return uniq
}
