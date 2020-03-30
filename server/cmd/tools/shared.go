package tools

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

func getFullPath(input string, err error) (string, error) {
	if err != nil {
		return "", err
	}

	path := input
	if !utils.IsDir(path) {
		return "", errors.New(fmt.Sprintf("%s is not a directory", path))
	}

	return strings.TrimRight(path, "/"), nil
}

func writeToFile(path string, data []byte) error {
	err := ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
