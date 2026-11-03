// Package engine renders a template (base + modules) into a directory.
package engine

import (
	"bytes"
	"strings"
	"text/template"
)

var funcs = template.FuncMap{
	"has":   has,
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
	"kebab": Slug,
	"snake": func(s string) string { return strings.ReplaceAll(Slug(s), "-", "_") },
}

// has checks membership in a list (the result of a multiselect).
func has(list any, v string) bool {
	switch l := list.(type) {
	case []string:
		for _, x := range l {
			if x == v {
				return true
			}
		}
	case []any:
		for _, x := range l {
			if s, ok := x.(string); ok && s == v {
				return true
			}
		}
	}
	return false
}

// Slug converts "My RAG Project!" to "my-rag-project".
func Slug(s string) string {
	var b strings.Builder
	prev := '-'
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prev = r
		} else if prev != '-' {
			b.WriteByte('-')
			prev = '-'
		}
	}
	return strings.Trim(b.String(), "-")
}

// Render executes text as a template. An unknown key is an error,
// which catches typos in manifests.
func Render(name, text string, data any) (string, error) {
	t, err := template.New(name).Funcs(funcs).Option("missingkey=error").Parse(text)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Eval evaluates a `when` condition: empty means true; otherwise the rendered
// value must be "true".
func Eval(cond string, data any) (bool, error) {
	if strings.TrimSpace(cond) == "" {
		return true, nil
	}
	out, err := Render("when", cond, data)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "true", nil
}
