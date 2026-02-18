package gen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"github.com/artem328/depquery-go/internal/gen/generator"
	"github.com/artem328/depquery-go/internal/gen/parser"
)

type Config struct {
	SchemaFile string
	Package    string
	OutputDir  string
	Workers    int
}

type multiErr []error

func (e multiErr) Error() string {
	msgs := make([]string, len(e))
	for i := 0; i < len(e); i++ {
		msgs[i] = e[i].Error()
	}

	return strings.Join(msgs, "\n")
}

func Generate(ctx context.Context, conf Config) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	out, err := outputDir(wd, conf.OutputDir)
	if err != nil {
		return err
	}

	abs, rel, err := schemaFilePaths(wd, out, conf.SchemaFile)
	if err != nil {
		return err
	}

	p, err := parser.FromFile(abs)
	if err != nil {
		return fmt.Errorf("parse schema file: %w", err)
	}

	schema, errs := p.Parse()
	if len(errs) > 0 {
		return multiErr(errs)
	}

	workers := conf.Workers
	if workers < 1 {
		workers = runtime.NumCPU()
	}

	return generator.Generate(ctx, workers, schema, pkgName(out, conf.Package), out, rel)
}

func schemaFilePaths(wd, dest, f string) (abs, rel string, err error) {
	abs = f

	if !filepath.IsAbs(abs) {
		abs, err = filepath.Abs(filepath.Join(wd, f))
		if err != nil {
			return
		}
	}

	rel, err = filepath.Rel(dest, abs)
	if err != nil {
		return
	}

	return
}

func outputDir(wd, d string) (string, error) {
	abs := d

	if filepath.IsAbs(abs) {
		return abs, nil
	}

	var err error

	abs, err = filepath.Abs(filepath.Join(wd, d))
	if err != nil {
		return "", nil
	}

	return abs, nil
}

func pkgName(outputDir, pkg string) string {
	if pkg != "" {
		return pkg
	}

	return sanitizePackageName(strings.ToLower(filepath.Base(outputDir)))
}

func sanitizePackageName(input string) string {
	var b strings.Builder

	// Ensure the first character is a letter (or prefix with "pkg")
	runes := []rune(input)
	if len(runes) == 0 || !unicode.IsLetter(runes[0]) {
		b.WriteString("pkg")
	}

	for _, r := range runes {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune('_') // replace invalid chars with _
		}
	}

	return b.String()
}
