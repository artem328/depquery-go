package generator

import (
	"strings"
	"sync"
	"unicode"
)

type nameKey struct {
	ID           string
	EntityName   string
	VariantName  string
	RelationName string
}

func nKey(id, e, v, r string) nameKey {
	return nameKey{ID: id, EntityName: e, VariantName: v, RelationName: r}
}

type nameIndex struct {
	mu  sync.RWMutex
	idx map[nameKey]string
}

func newNameIndex() *nameIndex {
	return &nameIndex{idx: make(map[nameKey]string)}
}

func (i *nameIndex) getOrCreate(k nameKey, create func() string) string {
	i.mu.RLock()
	if n, ok := i.idx[k]; ok {
		i.mu.RUnlock()

		return n
	}
	i.mu.RUnlock()

	i.mu.Lock()
	defer i.mu.Unlock()

	if n, ok := i.idx[k]; ok {
		return n
	}

	n := create()

	i.idx[k] = n

	return n
}

type sanitizeMode uint8

const (
	sanitizeRaw sanitizeMode = iota
	sanitizeRawCapitalized
	sanitizeUnexported
	sanitizeExported
)

//nolint:cyclop // probably refactor this
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
