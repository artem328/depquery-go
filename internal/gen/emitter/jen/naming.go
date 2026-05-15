package jen

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/artem328/depquery-go/internal/gen/plan"
)

type naming struct {
	Compiler         compilerNaming
	Resolver         resolverNaming
	Builder          builderNaming
	Relation         relationNaming
	Candidate        candidateNaming
	BuildContext     buildContextNaming
	Plan             planNaming
	PrefetchResolver prefetchResolverNaming
	EntityPrefetcher entityPrefetcherNaming
	State            stateNaming
	Instance         instanceNaming
	FetchContext     fetchContextNaming
}

func (n *naming) warmUp(p plan.Plan) {
	n.Compiler.warmUp(p)
	n.Resolver.warmUp(p)
	n.Builder.warmUp(p)
	n.Relation.warmUp(p)
	n.Candidate.warmUp(p)
	n.BuildContext.warmUp(p)
	n.Plan.warmUp(p)
	n.PrefetchResolver.warmUp(p)
	n.EntityPrefetcher.warmUp(p)
	n.State.warmUp(p)
	n.Instance.warmUp(p)
	n.FetchContext.warmUp(p)
}

func pkgName(outputDir, pkg string) string {
	if pkg != "" {
		return sanitizePackageName(pkg)
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

type sanitizeMode uint8

const (
	sanitizeRaw sanitizeMode = iota
	sanitizeRawCapitalized
	sanitizeUnexported
	sanitizeExported
)

func sanitizeID(raw string, mode sanitizeMode) string {
	if raw == "" {
		return ""
	}

	var b strings.Builder

	b.Grow(len(raw) + 1)

	first := true
	isRaw := mode == sanitizeRaw || mode == sanitizeRawCapitalized
	capitalize := mode == sanitizeExported || mode == sanitizeRawCapitalized

	for _, r := range raw {
		switch {
		case unicode.IsLetter(r):
		case unicode.IsDigit(r):
			if first && !isRaw {
				if capitalize {
					b.WriteRune('X')
				} else {
					b.WriteRune('x')
				}

				first = false
				capitalize = false
			}
		case r == '_':
			if first && !isRaw {
				continue
			}
		default:
			capitalize = true

			continue
		}

		if capitalize {
			r = unicode.ToUpper(r)
			capitalize = false
		} else if first && mode == sanitizeUnexported {
			r = unicode.ToLower(r)
		}

		b.WriteRune(r)

		first = false
	}

	return b.String()
}
