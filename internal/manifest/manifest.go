// Package manifest describes a template's manifest.yaml format.
package manifest

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Option struct {
	Value string `yaml:"value"`
	Label string `yaml:"label"`
}

// Question: type = select | multiselect | confirm | input.
// When is a template expression that must render "true" (empty means always).
type Question struct {
	ID      string   `yaml:"id"`
	Prompt  string   `yaml:"prompt"`
	Type    string   `yaml:"type"`
	Options []Option `yaml:"options"`
	Default any      `yaml:"default"`
	When    string   `yaml:"when"`
}

// Module is a template subdirectory copied over base/ when When is true.
type Module struct {
	Path string `yaml:"path"`
	When string `yaml:"when"`
}

// Dep is a dependency injected into {{ .Deps }} / {{ .DevDeps }}.
type Dep struct {
	Name string `yaml:"name"`
	Dev  bool   `yaml:"dev"`
	When string `yaml:"when"`
}

type EnvVar struct {
	Key         string `yaml:"key"`
	Value       string `yaml:"value"`
	Description string `yaml:"description"`
	Secret      bool   `yaml:"secret"`
	When        string `yaml:"when"`
}

// Hook is a command executed after generation (no shell: direct argv).
type Hook struct {
	Description string   `yaml:"description"`
	Cmd         []string `yaml:"cmd"`
	When        string   `yaml:"when"`
}

type Requirement struct {
	Cmd  string `yaml:"cmd"`
	When string `yaml:"when"`
}

type Manifest struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Category    string        `yaml:"category"`
	Requires    []Requirement `yaml:"requires"`
	Questions   []Question    `yaml:"questions"`
	Modules     []Module      `yaml:"modules"`
	Deps        []Dep         `yaml:"deps"`
	Env         []EnvVar      `yaml:"env"`
	Post        []Hook        `yaml:"post"`
	NextSteps   []string      `yaml:"next_steps"`
}

func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Manifest) Validate() error {
	if m.Name == "" || m.Category == "" {
		return fmt.Errorf("manifest: name and category are required")
	}
	seen := map[string]bool{}
	for _, q := range m.Questions {
		if q.ID == "" || seen[q.ID] {
			return fmt.Errorf("manifest: empty or duplicate question ID %q", q.ID)
		}
		seen[q.ID] = true
		switch q.Type {
		case "select", "multiselect":
			if len(q.Options) == 0 {
				return fmt.Errorf("manifest: question %q has no options", q.ID)
			}
		case "confirm", "input":
		default:
			return fmt.Errorf("manifest: unknown type %q for question %q", q.Type, q.ID)
		}
	}
	for _, mod := range m.Modules {
		if mod.Path == "" {
			return fmt.Errorf("manifest: module sans path")
		}
	}
	for _, h := range m.Post {
		if len(h.Cmd) == 0 {
			return fmt.Errorf("manifest: hook sans cmd")
		}
	}
	return nil
}
