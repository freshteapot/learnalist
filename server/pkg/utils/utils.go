package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func IsDir(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func CmdParsePathToFolder(key string, dir string) (string, error) {
	dir = strings.TrimRight(dir, "/")
	pathToFolder, _ := filepath.Abs(dir)

	if pathToFolder == "" {
		return "", fmt.Errorf("you might have forgotten to set the path for: %s", key)
	}

	if !IsDir(pathToFolder) {
		return "", fmt.Errorf("%s is not a directory", key)
	}

	return pathToFolder, nil
}

func PrettyPrintJSON(input []byte) string {
	var prettyJSON bytes.Buffer
	// Based on jq standard output
	json.Indent(&prettyJSON, input, "", "  ")
	return prettyJSON.String()
}

func IntArrayContains(items []int, find int) bool {
	for _, item := range items {
		if item == find {
			return true
		}
	}
	return false
}
func StringArrayContains(items []string, find string) bool {
	for _, item := range items {
		if item == find {
			return true
		}
	}
	return false
}

func StringArrayIndexOf(items []string, search string) int {
	for i, v := range items {
		if v == search {
			return i
		}
	}
	return -1
}

func StringArrayRemoveAtIndex(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
