package engine

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/abdellatifmht/Brik/internal/manifest"
)

type Options struct {
	Manifest  *manifest.Manifest
	Templates fs.FS // Template root (contains base/, modules/, manifest.yaml).
	Dest      string
	Data      map[string]any // Answers plus project_name, project_slug...
}

type Result struct {
	Files          []string
	MissingSecrets []string
}

func Generate(o Options) (*Result, error) {
	m, data := o.Manifest, o.Data

	if entries, err := os.ReadDir(o.Dest); err == nil && len(entries) > 0 {
		return nil, fmt.Errorf("%s already exists and is not empty", o.Dest)
	}

	// Dependencies computed from the selected options.
	deps, dev := []string{}, []string{}
	for _, d := range m.Deps {
		ok, err := Eval(d.When, data)
		if err != nil {
			return nil, fmt.Errorf("deps[%s].when: %w", d.Name, err)
		}
		if !ok {
			continue
		}
		if d.Dev {
			dev = append(dev, d.Name)
		} else {
			deps = append(deps, d.Name)
		}
	}
	data["Deps"], data["DevDeps"] = deps, dev

	roots := []string{"base"}
	for _, mod := range m.Modules {
		ok, err := Eval(mod.When, data)
		if err != nil {
			return nil, fmt.Errorf("modules[%s].when: %w", mod.Path, err)
		}
		if ok {
			roots = append(roots, mod.Path)
		}
	}

	res := &Result{}
	for _, root := range roots {
		if err := copyTree(o.Templates, root, o.Dest, data, res); err != nil {
			return nil, fmt.Errorf("%s: %w", root, err)
		}
	}
	if err := writeEnv(m, data, o.Dest, res); err != nil {
		return nil, err
	}
	return res, nil
}

func copyTree(src fs.FS, root, dest string, data map[string]any, res *Result) error {
	return fs.WalkDir(src, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, root), "/")
		if rel == "" {
			return nil
		}
		// Directory and file names are templates too.
		parts := strings.Split(rel, "/")
		for i, seg := range parts {
			out, err := Render(p, seg, data)
			if err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}
			parts[i] = out
		}
		outRel := path.Join(parts...)
		target := filepath.Join(dest, filepath.FromSlash(outRel))

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		raw, err := fs.ReadFile(src, p)
		if err != nil {
			return err
		}
		if strings.HasSuffix(target, ".tmpl") {
			out, err := Render(p, string(raw), data)
			if err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}
			if strings.TrimSpace(out) == "" {
				return nil // Ignore conditional files that render empty.
			}
			raw = []byte(out)
			target = strings.TrimSuffix(target, ".tmpl")
			outRel = strings.TrimSuffix(outRel, ".tmpl")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(target, ".sh") {
			mode = 0o755
		}
		if err := os.WriteFile(target, raw, mode); err != nil {
			return err
		}
		res.Files = append(res.Files, outRel)
		return nil
	})
}

func writeEnv(m *manifest.Manifest, data map[string]any, dest string, res *Result) error {
	var env, example strings.Builder
	for _, e := range m.Env {
		ok, err := Eval(e.When, data)
		if err != nil {
			return fmt.Errorf("env[%s].when: %w", e.Key, err)
		}
		if !ok {
			continue
		}
		val, err := Render(e.Key, e.Value, data)
		if err != nil {
			return fmt.Errorf("env[%s].value: %w", e.Key, err)
		}
		if e.Description != "" {
			fmt.Fprintf(&env, "# %s\n", e.Description)
			fmt.Fprintf(&example, "# %s\n", e.Description)
		}
		fmt.Fprintf(&env, "%s=%s\n", e.Key, val)
		if e.Secret {
			fmt.Fprintf(&example, "%s=\n", e.Key)
			if val == "" {
				res.MissingSecrets = append(res.MissingSecrets, e.Key)
			}
		} else {
			fmt.Fprintf(&example, "%s=%s\n", e.Key, val)
		}
	}
	if env.Len() == 0 {
		return nil
	}
	if err := os.WriteFile(filepath.Join(dest, ".env"), []byte(env.String()), 0o600); err != nil {
		return err
	}
	res.Files = append(res.Files, ".env")
	if err := os.WriteFile(filepath.Join(dest, ".env.example"), []byte(example.String()), 0o644); err != nil {
		return err
	}
	res.Files = append(res.Files, ".env.example")
	return nil
}
