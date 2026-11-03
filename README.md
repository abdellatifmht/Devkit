# Devkit

A single binary to generate projects, create `.env`, initialize Git, and check tools.

```bash
go mod tidy          # first time: generates go.sum
make build           # -> bin/devkit
./bin/devkit new     # interactive menu
```

## Commandes

| Command | Purpose |
|---|---|
| `devkit new [template]` | Create a project (interactive menu or automation flags) |
| `devkit list` | List templates by category |
| `devkit doctor` | Check git, Docker, Python, Node, Go, and Java and show install commands |

Mode non interactif :

```bash
devkit new rag --name mon-rag --set vector=qdrant --set llm=ollama --set tools=docker,pytest --yes
```

`rag` is a shortcut for `rag-python` (unambiguous prefix). `--set key=value` answers any manifest question.

In the interactive menu, `Shift+Tab` goes back to the previous question, including from project name to project type and from project type to category. This lets you change the selected template or project name before continuing. `Tab` or `Enter` advances to the next question.

## Included templates

| ID | Category | Description |
|---|---|---|
| `rag-python` | AI | RAG pipeline (Qdrant, Chroma, or pgvector + LLM) |
| `agent-python` | AI | LangGraph agent with tools and optional MCP server |
| `mlops-python` | AI | Training, MLflow tracking, and FastAPI prediction API |
| `fastapi-python` | Web | Typed REST API with SQLite or PostgreSQL |
| `nextjs` | Web | Next.js App Router, TypeScript, Tailwind, and optional Prisma |
| `spring-boot-react` | Web | Spring Boot API and React frontend with Vite |

Examples:

```bash
devkit new agent-python     --name mon-agent --set llm=openai --set tools=mcp,pytest --yes
devkit new mlops-python     --name mon-mlops --set tracking=mlflow --set serving=fastapi --yes
devkit new fastapi-python   --name mon-api   --set db=postgres --set tools=docker,pytest --yes
devkit new nextjs           --name mon-site  --set database=sqlite --set tools=docker,vitest --yes
devkit new spring-boot-react --name mon-app  --set database=postgres --set tools=docker,github-actions --yes
```

## Architecture

```
cmd/devkit/            entry point
internal/cli/          Cobra commands (new, list, doctor)
internal/manifest/     manifest.yaml structure and validation
internal/answers/      flags / defaults / questions -> template data
internal/prompt/       TUI (huh) implementing answers.Asker
internal/engine/       base + module rendering, .env generation
internal/catalog/      embedded template loading (go:embed)
internal/doctor/       tool detection + OS-specific hints
templates/<id>/        manifest.yaml + base/ + modules/
```

### Template anatomy

```
templates/rag-python/
  manifest.yaml   # questions, modules, deps, env, hooks post, next_steps
  base/           # always copied
  modules/<name>/  # copied over base when `when` is true
```

- File and directory names are templates: `src/{{.project_module}}/`.
- A `*.tmpl` file is rendered (suffix removed); other files are copied as-is.
- A `*.tmpl` file that renders empty is ignored, which is useful for conditional files.
- Available variables: answers (`.vector`, `.tools`, ...), `.project_name`, `.project_slug`, `.project_module`, `.Deps`, `.DevDeps`.
- Functions: `has <list> "value"`, `upper`, `lower`, `kebab`, `snake`.
- `when` is an expression that must evaluate to `true`: `'{{ eq .llm "ollama" }}'`, `'{{ has .tools "docker" }}'`.
- An unknown template key is an error (never a silent `<no value>`).

### Adding a template

1. Create `templates/<id>/` with `manifest.yaml`, `base/`, and `modules/`.
2. Add `all:<id>` to the `//go:embed` directive in `templates/embed.go`.
3. `make test`: the test generates **all combinations** of `select` questions (with and without tools), rejects `<no value>` and unrendered template actions, then checks Python syntax.
