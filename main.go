package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/bobg/errors"
	"golang.org/x/sync/errgroup"
)

const lxSigil = "lx:"

type Entry struct {
	Path      string
	LXMessage string
}

func main() {
	timeout := flag.Duration("timeout", time.Second, "Command timeout.")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)

	var entries []Entry
	var entriesLock sync.Mutex
	group.Go(func() error {
		return filepath.WalkDir(".", func(path string, dirEntry fs.DirEntry, _ error) error {
			info, err := dirEntry.Info()
			if err != nil {
				return errors.Wrapf(err, "getting dir entry info for %q", path)
			}

			if info.IsDir() {
				return nil
			}

			if !isExecutable(info.Mode()) {
				return nil
			}

			group.Go(func() error {
				if isUTF8, err := isUTF8(path); err != nil {
					return errors.Wrapf(err, "detecting charset for %q", path)
				} else if !isUTF8 {
					return nil
				}

				lxMessage, err := extractFirstLXMessage(path)
				if err != nil {
					return errors.Wrapf(err, "extracting lx message from %q", path)
				}

				entry := Entry{
					Path:      path,
					LXMessage: lxMessage,
				}

				entriesLock.Lock()
				entries = append(entries, entry)
				entriesLock.Unlock()
				return nil
			})

			return nil
		})
	})

	if err := group.Wait(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	slices.SortFunc(entries, func(a, b Entry) int { return strings.Compare(a.Path, b.Path) })
	for _, entry := range entries {
		fmt.Println(entry.Path, "", entry.LXMessage)
	}
}

func isUTF8(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, errors.Wrapf(err, "opening file %q", path)
	}

	defer func() {
		if closeErr := file.Close(); err != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	return utf8.ValidString(scanner.Text()), nil
}

func extractFirstLXMessage(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", errors.Wrapf(err, "opening file %q", path)
	}

	defer func() {
		if closeErr := file.Close(); err != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if _, lxMessage, found := strings.Cut(scanner.Text(), lxSigil); found {
			return strings.TrimSpace(lxMessage), nil
		}
	}

	return "", nil
}

func extractLXMessages(path string) (_ []string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "opening file %q", path)
	}

	defer func() {
		if closeErr := file.Close(); err != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	var lxMessages []string
	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		if _, lxMessage, found := strings.Cut(scanner.Text(), lxSigil); found {
			lxMessages = append(lxMessages, strings.TrimSpace(lxMessage))
		}
	}

	return lxMessages, nil
}

func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}
