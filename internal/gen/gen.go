package gen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/artem328/depquery-go/internal/gen/emitter"
	"github.com/artem328/depquery-go/internal/gen/parser"
	"github.com/artem328/depquery-go/internal/gen/plan"
	"github.com/artem328/depquery-go/internal/gen/schema"
	"github.com/artem328/depquery-go/internal/gen/semantic"
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

func Generate(conf Config) error {
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

	prs, err := parser.FromFile(abs)
	if err != nil {
		return fmt.Errorf("parse schema file: %w", err)
	}

	s, errs := prs.Parse()
	if len(errs) > 0 {
		return multiErr(errs)
	}

	if errs := schema.Validate(s); len(errs) > 0 {
		return multiErr(errs)
	}

	r := semantic.NewResolver(s)

	m, errs := r.Build()
	if len(errs) > 0 {
		return multiErr(errs)
	}

	pln := plan.NewPlanner(m)

	p, err := pln.Plan()
	if err != nil {
		return err
	}

	e := emitter.New(p)

	return e.Emit(out, conf.Package, rel)
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
