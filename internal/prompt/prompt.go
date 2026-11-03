// Package prompt implements answers.Asker with the huh TUI library.
package prompt

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/abdellatifmht/Brik/internal/answers"
	"github.com/abdellatifmht/Brik/internal/engine"
	"github.com/abdellatifmht/Brik/internal/manifest"
)

type TemplateGroup struct {
	Category string
	Options  []huh.Option[string]
}

type TemplateSelection struct {
	ID   string
	Name string
}

func run(f huh.Field) error {
	return huh.NewForm(huh.NewGroup(f)).Run()
}

func label(o manifest.Option) string {
	if o.Label != "" {
		return o.Label
	}
	return o.Value
}

func contains(l []string, v string) bool {
	for _, x := range l {
		if x == v {
			return true
		}
	}
	return false
}

// Ask asks one question according to its type.
func Ask(q manifest.Question, def any) (any, error) {
	switch q.Type {
	case "select":
		v := def.(string)
		opts := make([]huh.Option[string], len(q.Options))
		for i, o := range q.Options {
			opts[i] = huh.NewOption(label(o), o.Value)
		}
		err := run(huh.NewSelect[string]().Title(q.Prompt).Options(opts...).Value(&v))
		return v, err

	case "multiselect":
		defaults := def.([]string)
		v := append([]string(nil), defaults...)
		opts := make([]huh.Option[string], len(q.Options))
		for i, o := range q.Options {
			opts[i] = huh.NewOption(label(o), o.Value).Selected(contains(defaults, o.Value))
		}
		err := run(huh.NewMultiSelect[string]().Title(q.Prompt).Options(opts...).Value(&v))
		return v, err

	case "confirm":
		v := def.(bool)
		err := run(huh.NewConfirm().Title(q.Prompt).Value(&v))
		return v, err
	}
	v := def.(string)
	err := run(huh.NewInput().Title(q.Prompt).Value(&v))
	return v, err
}

// AskAll asks all questions in one form. Shift+Tab goes back to the previous
// question, while Enter/Tab advances to the next one.
func AskAll(questions []answers.FormQuestion, initial map[string]any) (map[string]any, error) {
	if len(questions) == 0 {
		return map[string]any{}, nil
	}

	values := make(map[string]any, len(initial))
	for k, v := range initial {
		values[k] = v
	}

	getters := make(map[string]func() any, len(questions))
	groups := make([]*huh.Group, 0, len(questions))
	for _, item := range questions {
		def := item.Default
		if def == nil {
			def = answers.Default(item.Question)
		}
		field, get := field(item.Question, def)
		getters[item.Question.ID] = get

		when := item.Question.When
		groups = append(groups, huh.NewGroup(field).WithHideFunc(func() bool {
			ok, err := engine.Eval(when, snapshot(values, getters))
			return err != nil || !ok
		}))
	}

	// Validate conditions before displaying the form so an unknown key produces
	// a clear error instead of an incomplete form.
	for _, item := range questions {
		if _, err := engine.Eval(item.Question.When, snapshot(values, getters)); err != nil {
			return nil, fmt.Errorf("question %q: %w", item.Question.ID, err)
		}
	}

	if err := huh.NewForm(groups...).Run(); err != nil {
		return nil, err
	}
	out := make(map[string]any, len(getters))
	for id, get := range getters {
		out[id] = get()
	}
	return out, nil
}

func field(q manifest.Question, def any) (huh.Field, func() any) {
	switch q.Type {
	case "select":
		v := def.(string)
		opts := make([]huh.Option[string], len(q.Options))
		for i, o := range q.Options {
			opts[i] = huh.NewOption(label(o), o.Value)
		}
		f := huh.NewSelect[string]().Title(q.Prompt).Options(opts...).Value(&v)
		return f, func() any { return v }

	case "multiselect":
		defaults := def.([]string)
		v := append([]string(nil), defaults...)
		opts := make([]huh.Option[string], len(q.Options))
		for i, o := range q.Options {
			opts[i] = huh.NewOption(label(o), o.Value).Selected(contains(defaults, o.Value))
		}
		f := huh.NewMultiSelect[string]().Title(q.Prompt).Options(opts...).Value(&v)
		return f, func() any { return v }

	case "confirm":
		v := def.(bool)
		f := huh.NewConfirm().Title(q.Prompt).Value(&v)
		return f, func() any { return v }
	}

	v := def.(string)
	f := huh.NewInput().Title(q.Prompt).Value(&v)
	return f, func() any { return v }
}

func snapshot(initial map[string]any, getters map[string]func() any) map[string]any {
	out := make(map[string]any, len(initial)+len(getters))
	for k, v := range initial {
		out[k] = v
	}
	for k, get := range getters {
		out[k] = get()
	}
	return out
}

// Input asks for a simple free-form value, such as a project name.
func Input(title, def string) (string, error) {
	v := def
	err := run(huh.NewInput().Title(title).Value(&v))
	return v, err
}

// Select chooses a string from a list, such as a category or template.
func Select(title string, values, labels []string) (string, error) {
	opts := make([]huh.Option[string], len(values))
	for i := range values {
		opts[i] = huh.NewOption(labels[i], values[i])
	}
	var v string
	err := run(huh.NewSelect[string]().Title(title).Options(opts...).Value(&v))
	return v, err
}

// SelectTemplate presents category and template selection in one navigable
// form. Shift+Tab from a template returns to the category selection.
func SelectTemplate(groups []TemplateGroup) (string, error) {
	selection, err := SelectTemplateAndName(groups, "")
	if err != nil {
		return "", err
	}
	return selection.ID, nil
}

// SelectTemplateAndName presents template selection and project naming in one
// navigable form. Shift+Tab from the name returns to the template selection.
func SelectTemplateAndName(groups []TemplateGroup, initialName string) (TemplateSelection, error) {
	if len(groups) == 0 {
		return TemplateSelection{}, fmt.Errorf("no template categories available")
	}

	category := groups[0].Category
	values := make(map[string]*string, len(groups))
	formGroups := make([]*huh.Group, 0, len(groups)+1)
	categories := make([]huh.Option[string], len(groups))
	for i, group := range groups {
		categories[i] = huh.NewOption(group.Category, group.Category)
		if len(group.Options) > 0 {
			value := group.Options[0].Value
			values[group.Category] = &value
		}
	}

	categoryField := huh.NewSelect[string]().
		Title("Category").
		Options(categories...).
		Value(&category)
	formGroups = append(formGroups, huh.NewGroup(categoryField))

	for _, group := range groups {
		group := group
		templateField := huh.NewSelect[string]().
			Title("Project type").
			Options(group.Options...).
			Value(values[group.Category])
		formGroups = append(formGroups, huh.NewGroup(templateField).WithHideFunc(func() bool {
			return category != group.Category
		}))
	}

	templateID := *values[category]
	name := initialName
	if name == "" {
		name = "my-" + templateID
	}
	nameField := huh.NewInput().Title("Project name").Value(&name)
	formGroups = append(formGroups, huh.NewGroup(nameField))

	if err := huh.NewForm(formGroups...).Run(); err != nil {
		return TemplateSelection{}, err
	}
	return TemplateSelection{ID: *values[category], Name: name}, nil
}
