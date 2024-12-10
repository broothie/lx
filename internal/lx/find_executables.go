package lx

import (
	"bufio"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/bobg/errors"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

const lxSigil = "lx:"

type Entry struct {
	Path    string
	Message string
}

func FindExecutables(ctx context.Context, rootPath string, skipDirs []string) ([]Entry, error) {
	var entries []Entry
	var entriesLock sync.Mutex

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return filepath.WalkDir(rootPath, func(path string, dirEntry fs.DirEntry, _ error) error {
			info, err := dirEntry.Info()
			if err != nil {
				return errors.Wrapf(err, "getting dir entry info for %q", path)
			}

			if info.IsDir() {
				if lo.Contains(skipDirs, path) {
					return fs.SkipDir
				}

				return nil
			}

			if !isExecutable(info.Mode()) {
				return nil
			}

			group.Go(func() error {
				if isUTF8, err := fileIsUTF8(path); err != nil {
					return errors.Wrapf(err, "detecting charset for %q", path)
				} else if !isUTF8 {
					return nil
				}

				lxMessage, err := extractFirstLXMessage(path)
				if err != nil {
					return errors.Wrapf(err, "extracting lx message from %q", path)
				}

				entry := Entry{
					Path:    path,
					Message: lxMessage,
				}

				entriesLock.Lock()
				entries = append(entries, entry)
				entriesLock.Unlock()
				return nil
			})

			return nil
		})
	})

	return entries, group.Wait()
}

func fileIsUTF8(path string) (bool, error) {
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

func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}
