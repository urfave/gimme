package internal

import (
	"os"
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
