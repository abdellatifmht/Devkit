package cli

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/abdellatifmht/Brik/internal/answers"
	"github.com/abdellatifmht/Brik/internal/catalog"
	"github.com/abdellatifmht/Brik/internal/doctor"
	"github.com/abdellatifmht/Brik/internal/engine"
	"github.com/abdellatifmht/Brik/internal/prompt"
)

func newCmd() *cobra.Command {
	var (
		name, out string
		sets      []string
		yes       bool
		noHooks   bool
	)
	cmd := &cobra.Command{
		Use:   "new [template]",
		Short: "Create a project from a template (interactive menu if none is given)",
		Example: `  devkit new
  devkit new rag-python --name mon-rag --set vector=qdrant --set llm=ollama --yes`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runNew(args, name, out, sets, yes, noHooks)
			if errors.Is(err, huh.ErrUserAborted) {
				fmt.Println("Canceled.")
				return nil
			}
			return err
		},
	}
	f := cmd.Flags()
	f.StringVarP(&name, "name", "n", "", "project name")
	f.StringVarP(&out, "out", "o", ".", "destination parent directory")
	f.StringArrayVar(&sets, "set", nil, "answer a question: --set key=value (repeatable)")
	f.BoolVarP(&yes, "yes", "y", false, "accept default values (non-interactive mode)")
	f.BoolVar(&noHooks, "no-hooks", false, "do not run post-generation commands")
	return cmd
}

func runNew(args []string, name, out string, sets []string, yes, noHooks bool) error {
	entries, err := catalog.Load()
	if err != nil {
		return err
	}

	// 1. Choose the template
	var entry *catalog.Entry
	switch {
	case len(args) == 1:
		if entry = catalog.Find(entries, args[0]); entry == nil {
			return fmt.Errorf("template %q not found (see `devkit list`)", args[0])
		}
	case yes:
		return errors.New("--yes requires a template")
	case name == "":
		groups := templateGroups(entries)
		selection, err := prompt.SelectTemplateAndName(groups, "")
		if err != nil {
			return err
		}
		entry = catalog.Find(entries, selection.ID)
		name = selection.Name
	default:
		if entry, err = pickTemplate(entries); err != nil {
			return err
		}
	}
	m := entry.Manifest

	// 2. Project name
	if name == "" {
		if yes {
			name = "my-" + entry.ID
		} else if name, err = prompt.Input("Project name", "my-"+entry.ID); err != nil {
			return err
		}
	}
	slug := engine.Slug(name)
	if slug == "" {
		return errors.New("invalid project name")
	}
	dest := filepath.Join(out, slug)

	// 3. Answers
	preset, err := parseSets(sets)
	if err != nil {
		return err
	}
	base := map[string]any{"project_name": name, "project_slug": slug, "project_module": strings.ReplaceAll(slug, "-", "_")}
	data, err := answers.ResolveWithForm(m, base, preset, yes, prompt.Ask, prompt.AskAll)
	if err != nil {
		return err
	}

	// 4. Pre-check required tools
	for _, r := range m.Requires {
		if ok, _ := engine.Eval(r.When, data); ok && !doctor.Has(r.Cmd) {
			fmt.Printf("⚠ %s not found → %s\n", r.Cmd, doctor.Hint(r.Cmd))
		}
	}

	// 5. Generate
	res, err := engine.Generate(engine.Options{Manifest: m, Templates: entry.FS, Dest: dest, Data: data})
	if err != nil {
		return err
	}

	// 6. Hooks (Git init, etc.)
	if !noHooks {
		for _, h := range m.Post {
			if ok, err := engine.Eval(h.When, data); err != nil {
				return err
			} else if !ok {
				continue
			}
			argv := make([]string, len(h.Cmd))
			for i, a := range h.Cmd {
				if argv[i], err = engine.Render("hook", a, data); err != nil {
					return err
				}
			}
			c := exec.Command(argv[0], argv[1:]...)
			c.Dir = dest
			if b, err := c.CombinedOutput(); err != nil {
				fmt.Printf("⚠ %s failed: %v\n%s", strings.Join(argv, " "), err, b)
			}
		}
	}

	// 7. Summary
	fmt.Printf("\n✔ Project created in %s (%d files)\n", dest, len(res.Files))
	if len(res.MissingSecrets) > 0 {
		fmt.Printf("⚠ Secrets to add to .env: %s\n", strings.Join(res.MissingSecrets, ", "))
	}
	if len(m.NextSteps) > 0 {
		fmt.Println("\nNext steps:")
		fmt.Printf("  cd %s\n", dest)
		for _, s := range m.NextSteps {
			line, err := engine.Render("next", s, data)
			if err != nil {
				return err
			}
			if strings.TrimSpace(line) != "" {
				fmt.Println("  " + line)
			}
		}
	}
	return nil
}

func pickTemplate(entries []catalog.Entry) (*catalog.Entry, error) {
	groups := templateGroups(entries)
	id, err := prompt.SelectTemplate(groups)
	if err != nil {
		return nil, err
	}
	return catalog.Find(entries, id), nil
}

func templateGroups(entries []catalog.Entry) []prompt.TemplateGroup {
	cats := catalog.Categories(entries)
	groups := make([]prompt.TemplateGroup, 0, len(cats))
	for _, cat := range cats {
		options := make([]huh.Option[string], 0)
		for _, e := range entries {
			if e.Manifest.Category != cat {
				continue
			}
			options = append(options, huh.NewOption(
				fmt.Sprintf("%s — %s", e.Manifest.Name, e.Manifest.Description),
				e.ID,
			))
		}
		groups = append(groups, prompt.TemplateGroup{Category: cat, Options: options})
	}
	return groups
}

func parseSets(sets []string) (map[string]string, error) {
	out := map[string]string{}
	for _, s := range sets {
		k, v, ok := strings.Cut(s, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("--set %q: expected key=value", s)
		}
		out[k] = v
	}
	return out, nil
}
