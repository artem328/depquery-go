package emitter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/artem328/depquery-go/internal/gen/emitter/jen"
	"github.com/artem328/depquery-go/internal/gen/plan"
)

type Emitter struct {
	plan plan.Plan
}

func New(p plan.Plan) *Emitter {
	return &Emitter{plan: p}
}

func (e *Emitter) Emit(dest, pkg, schemaLocation string) error {
	r := jen.NewRenderer(e.plan, dest, pkg, schemaLocation)

	bbuf, err := r.RenderBuilders()
	if err != nil {
		return err
	}

	dqbuf, err := r.RenderDepquery()
	if err != nil {
		return err
	}

	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("clear directory: %w", err)
	}

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(dest, "builder.go"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := io.Copy(f, bbuf); err != nil {
		return err
	}

	f, err = os.OpenFile(filepath.Join(dest, "depquery.go"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := io.Copy(f, dqbuf); err != nil {
		return err
	}

	return nil
}
