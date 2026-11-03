// Package answers transforms manifest questions into answers:
// flags (--set), default values (--yes), or interactive prompts (Asker).
package answers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/abdellatifmht/Brik/internal/engine"
	"github.com/abdellatifmht/Brik/internal/manifest"
)

// Asker asks a human one question; def is the typed default value.
type Asker func(q manifest.Question, def any) (any, error)

// FormQuestion is an interactive question with its default value.
type FormQuestion struct {
	Question manifest.Question
	Default  any
}

// FormAsker asks all questions in one form. This allows the user to go back to
// a previous question (Shift+Tab in the TUI) before submitting.
// initial contains known values: presets, base answers, and defaults.
type FormAsker func(questions []FormQuestion, initial map[string]any) (map[string]any, error)

func Zero(q manifest.Question) any {
	switch q.Type {
	case "multiselect":
		return []string{}
	case "confirm":
		return false
	}
	return ""
}

func Default(q manifest.Question) any {
	switch q.Type {
	case "select":
		if s, ok := q.Default.(string); ok {
			return s
		}
		return q.Options[0].Value
	case "multiselect":
		out := []string{}
		if l, ok := q.Default.([]any); ok {
			for _, x := range l {
				out = append(out, fmt.Sprint(x))
			}
		}
		return out
	case "confirm":
		b, _ := q.Default.(bool)
		return b
	}
	s, _ := q.Default.(string)
	return s
}

// Resolve returns the template data map. base contains already-known variables
// (project_name...). Questions whose `when` is false keep their zero value so
// templates can always reference them.
func Resolve(m *manifest.Manifest, base map[string]any, preset map[string]string, yes bool, ask Asker) (map[string]any, error) {
	return ResolveWithForm(m, base, preset, yes, ask, nil)
}

// ResolveWithForm is like Resolve, but uses form (when provided) to ask all
// questions in one navigable form. Questions supplied through --set are fixed
// and are not asked.
func ResolveWithForm(m *manifest.Manifest, base map[string]any, preset map[string]string, yes bool, ask Asker, form FormAsker) (map[string]any, error) {
	data := map[string]any{}
	for k, v := range base {
		data[k] = v
	}
	ids := map[string]bool{}
	for _, q := range m.Questions {
		data[q.ID] = Zero(q)
		ids[q.ID] = true
	}
	for k := range preset {
		if !ids[k] {
			return nil, fmt.Errorf("--set %s: unknown key for this template (valid keys: %s)", k, strings.Join(keys(m), ", "))
		}
	}

	if !yes && form != nil {
		return resolveForm(m, data, preset, form)
	}

	for _, q := range m.Questions {
		ok, err := engine.Eval(q.When, data)
		if err != nil {
			return nil, fmt.Errorf("question %q: %w", q.ID, err)
		}
		if !ok {
			continue
		}
		switch raw, set := preset[q.ID]; {
		case set:
			v, err := parse(q, raw)
			if err != nil {
				return nil, err
			}
			data[q.ID] = v
		case yes || ask == nil:
			data[q.ID] = Default(q)
		default:
			v, err := ask(q, Default(q))
			if err != nil {
				return nil, err
			}
			data[q.ID] = v
		}
	}
	return data, nil
}

func resolveForm(m *manifest.Manifest, data map[string]any, preset map[string]string, form FormAsker) (map[string]any, error) {
	// Pre-fill answers so `when` conditions can be evaluated as soon as the form
	// opens.
	for _, q := range m.Questions {
		if raw, set := preset[q.ID]; set {
			v, err := parse(q, raw)
			if err != nil {
				return nil, err
			}
			data[q.ID] = v
			continue
		}
		data[q.ID] = Default(q)
	}

	questions := make([]FormQuestion, 0, len(m.Questions))
	for _, q := range m.Questions {
		if _, set := preset[q.ID]; set {
			continue
		}
		questions = append(questions, FormQuestion{Question: q, Default: Default(q)})
	}
	if len(questions) > 0 {
		initial := make(map[string]any, len(data))
		for k, v := range data {
			initial[k] = v
		}
		got, err := form(questions, initial)
		if err != nil {
			return nil, err
		}
		for k, v := range got {
			data[k] = v
		}
	}

	// Clear an answer that became inactive after navigating backward, matching
	// the question-by-question flow.
	for _, q := range m.Questions {
		ok, err := engine.Eval(q.When, data)
		if err != nil {
			return nil, fmt.Errorf("question %q: %w", q.ID, err)
		}
		if !ok {
			data[q.ID] = Zero(q)
		}
	}
	return data, nil
}

func keys(m *manifest.Manifest) []string {
	var out []string
	for _, q := range m.Questions {
		out = append(out, q.ID)
	}
	return out
}

func validOption(q manifest.Question, v string) bool {
	for _, o := range q.Options {
		if o.Value == v {
			return true
		}
	}
	return false
}

func parse(q manifest.Question, raw string) (any, error) {
	switch q.Type {
	case "select":
		if !validOption(q, raw) {
			return nil, fmt.Errorf("--set %s=%s: invalid value", q.ID, raw)
		}
		return raw, nil
	case "multiselect":
		out := []string{}
		for _, v := range strings.Split(raw, ",") {
			if v = strings.TrimSpace(v); v == "" {
				continue
			}
			if !validOption(q, v) {
				return nil, fmt.Errorf("--set %s: %q is not a valid option", q.ID, v)
			}
			out = append(out, v)
		}
		return out, nil
	case "confirm":
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("--set %s=%s: expected a boolean", q.ID, raw)
		}
		return b, nil
	}
	return raw, nil
}
