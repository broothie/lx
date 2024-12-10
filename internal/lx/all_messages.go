package lx

import (
	"bufio"
	"os"
	"strings"

	"github.com/bobg/errors"
)

func AllMessages(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", errors.Wrapf(err, "opening file %q", path)
	}

	defer func() {
		if closeErr := file.Close(); err != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	var messages []string
	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		if _, lxMessage, found := strings.Cut(scanner.Text(), lxSigil); found {
			messages = append(messages, strings.TrimSpace(lxMessage))
		}
	}

	return strings.Join(messages, "\n"), nil
}
