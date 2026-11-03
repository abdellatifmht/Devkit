// Package catalog loads embedded templates.
package catalog

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/abdellatifmht/Brik/internal/manifest"
	"github.com/abdellatifmht/Brik/templates"
)

type Entry struct {
	ID       string
	Manifest *manifest.Manifest
	FS       fs.FS
}

func Load() ([]Entry, error) {
	dirs, err := fs.ReadDir(templates.FS, ".")
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		raw, err := fs.ReadFile(templates.FS, d.Name()+"/manifest.yaml")
		if err != nil {
			continue
		}
		m, err := manifest.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("template %s: %w", d.Name(), err)
		}
		sub, err := fs.Sub(templates.FS, d.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, Entry{ID: d.Name(), Manifest: m, FS: sub})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Find accepts an exact ID ("rag-python") or an unambiguous prefix ("rag").
func Find(entries []Entry, id string) *Entry {
	var matches []*Entry
	for i := range entries {
		if entries[i].ID == id {
			return &entries[i]
		}
		if strings.HasPrefix(entries[i].ID, id+"-") {
			matches = append(matches, &entries[i])
		}
	}
	if len(matches) == 1 {
		return matches[0]
	}
	return nil
}

func Categories(entries []Entry) []string {
	seen := map[string]bool{}
	var cats []string
	for _, e := range entries {
		if !seen[e.Manifest.Category] {
			seen[e.Manifest.Category] = true
			cats = append(cats, e.Manifest.Category)
		}
	}
	sort.Strings(cats)
	return cats
}
