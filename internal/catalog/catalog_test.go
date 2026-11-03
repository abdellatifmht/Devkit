package catalog_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abdellatifmht/Brik/internal/answers"
	"github.com/abdellatifmht/Brik/internal/catalog"
	"github.com/abdellatifmht/Brik/internal/engine"
	"github.com/abdellatifmht/Brik/internal/manifest"
)

// toolValues returns every option for the `tools` question, or "" when the
// template does not have one. This avoids requiring a `tools` question in
// every template.
func toolValues(m *manifest.Manifest) string {
	for _, q := range m.Questions {
		if q.ID != "tools" {
			continue
		}
		values := make([]string, 0, len(q.Options))
		for _, o := range q.Options {
			values = append(values, o.Value)
		}
		return strings.Join(values, ",")
	}
	return ""
}

// combos returns the Cartesian product of all `select` question options.
func combos(m *manifest.Manifest) []map[string]string {
	out := []map[string]string{{}}
	for _, q := range m.Questions {
		if q.Type != "select" {
			continue
		}
		var next []map[string]string
		for _, base := range out {
			for _, o := range q.Options {
				c := map[string]string{q.ID: o.Value}
				for k, v := range base {
					c[k] = v
				}
				next = append(next, c)
			}
		}
		out = next
	}
	return out
}

func TestAllTemplatesRenderInEveryCombination(t *testing.T) {
	entries, err := catalog.Load()
	if err != nil || len(entries) == 0 {
		t.Fatalf("empty or invalid catalog: %v", err)
	}
	_, pyErr := exec.LookPath("python3")

	for _, e := range entries {
		toolStates := []string{""}
		if all := toolValues(e.Manifest); all != "" {
			toolStates = append(toolStates, all)
		}
		for _, preset := range combos(e.Manifest) {
			for _, tools := range toolStates {
				p := map[string]string{}
				if tools != "" {
					p["tools"] = tools
				}
				for k, v := range preset {
					p[k] = v
				}
				name := fmt.Sprintf("%s/%v/tools=%q", e.ID, preset, tools)
				t.Run(name, func(t *testing.T) {
					base := map[string]any{"project_name": "Mon RAG", "project_slug": "mon-rag", "project_module": "mon_rag"}
					data, err := answers.Resolve(e.Manifest, base, p, true, nil)
					if err != nil {
						t.Fatal(err)
					}
					dest := filepath.Join(os.Getenv("DEVKIT_OUT"), strings.NewReplacer("/", "_", " ", "", "[", "", "]", "", "\"", "", ":", "-", ",", "+").Replace(name))
					if os.Getenv("DEVKIT_OUT") == "" {
						dest = filepath.Join(t.TempDir(), "mon-rag")
					}
					res, err := engine.Generate(engine.Options{Manifest: e.Manifest, Templates: e.FS, Dest: dest, Data: data})
					if err != nil {
						t.Fatal(err)
					}
					if err := filepath.Walk(dest, func(path string, info os.FileInfo, walkErr error) error {
						if walkErr != nil {
							return walkErr
						}
						if info.IsDir() {
							return nil
						}
						b, err := os.ReadFile(path)
						if err != nil {
							return err
						}
						if strings.Contains(string(b), "<no value>") {
							t.Errorf("%s contains <no value>", path)
						}
						if strings.Contains(string(b), "{{") {
							t.Errorf("%s contains an unrendered template action", path)
						}
						if pyErr == nil && strings.HasSuffix(path, ".py") {
							if out, err := exec.Command("python3", "-m", "py_compile", path).CombinedOutput(); err != nil {
								t.Errorf("%s: invalid Python syntax: %s", path, out)
							}
						}
						return nil
					}); err != nil {
						t.Fatal(err)
					}
					if _, err := os.Stat(filepath.Join(dest, ".env")); err != nil {
						t.Errorf(".env is missing")
					}
					if len(res.Files) < 8 {
						t.Errorf("too few generated files: %v", res.Files)
					}
				})
			}
		}
	}
}

func TestSlug(t *testing.T) {
	if got := engine.Slug("  Mon Projet RAG! "); got != "mon-projet-rag" {
		t.Fatalf("got %q", got)
	}
}

func TestUnknownSetKeyIsRejected(t *testing.T) {
	entries, _ := catalog.Load()
	_, err := answers.Resolve(entries[0].Manifest, nil, map[string]string{"nope": "x"}, true, nil)
	if err == nil {
		t.Fatal("expected an error for an unknown key")
	}
}
