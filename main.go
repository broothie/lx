package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bobg/errors"
	"github.com/broothie/lx/internal/lx"
)

var commaSplitter = regexp.MustCompile(`\s*,\s*`)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	timeout := flag.Duration("timeout", time.Second, "Command timeout.")
	rootPath := flag.String("root", ".", "Root path to start search.")
	skipDirs := flag.String("skip-dirs", strings.Join(lx.DefaultSkipDirs(), ","), "Directories to skip during search.")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	path := flag.Arg(0)
	if path == "" {
		entries, err := lx.FindExecutables(ctx, *rootPath, commaSplitter.Split(*skipDirs, -1))
		if err != nil {
			return errors.Wrap(err, "finding executables")
		}

		rows := [][]string{{"PATH", "INFO"}}
		for _, entry := range entries {
			rows = append(rows, []string{entry.Path, entry.Message})
		}

		return writeTable(os.Stdout, rows)
	}

	message, err := lx.AllMessages(path)
	if err != nil {
		return errors.Wrap(err, "getting all messages")
	}

	fmt.Println(path)
	fmt.Println(message)
	return nil
}

func writeTable(w io.Writer, rows [][]string) error {
	table := tabwriter.NewWriter(w, 0, 1, 2, ' ', 0)
	for _, row := range rows {
		if _, err := fmt.Fprintln(table, strings.Join(row, "\t")); err != nil {
			return errors.Wrap(err, "writing row")
		}
	}

	if err := table.Flush(); err != nil {
		return errors.Wrap(err, "flushing table")
	}

	return nil
}
