package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/artem328/depquery-go/internal/gen"
)

func main() {
	os.Exit(run())
}

func run() int {
	var conf gen.Config

	flag.StringVar(&conf.SchemaFile, "schema", "", "path to a file containing schema definition")
	flag.StringVar(
		&conf.Package,
		"package",
		"",
		"package name for generated files (by default inferred from output directory)",
	)
	flag.StringVar(
		&conf.OutputDir,
		"output-dir",
		"./depquery",
		"path to output directory. NOTE: the directory will be cleared before new files are emitted",
	)
	flag.IntVar(&conf.Workers, "workers", runtime.NumCPU(), "number of concurrent workers")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := gen.Generate(ctx, conf); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		return 1
	}

	return 0
}
