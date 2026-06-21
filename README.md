# glab-dep

A GitLab CLI extension that streamlines the review and merge workflow for automated dependency update merge requests (MRs).

It delegates all GitLab access to the [`glab`](https://gitlab.com/gitlab-org/cli) CLI, so it never stores a token itself — authentication is whatever `glab` is already configured with.

## Features

- 🖥️ **Interactive TUI**: Full-featured terminal UI with keyboard navigation and live settings adjustment
- 📋 **List** dependency MRs by label/author with clean table output
- 📦 **Group** MRs by `package@version` for easier batched review
- ✅ **Bulk approve** all MRs for a chosen group via `glab mr approve`
- 🚀 **Bulk merge** per group via `glab mr merge`, optionally using GitLab's native auto-merge (merge when the pipeline succeeds)
- 🏢 **Multi-project support**: Target specific projects or an entire group/subgroup
- 🔄 Works out-of-the-box with **Renovate**
- 🎨 **Multiple output formats**: Human-readable tables or JSON
- ⚙️ **Configuration support**: Save default projects, author, and custom patterns via `glab config`
- 🎯 **Custom patterns**: Define your own MR title patterns for grouping

## Installation

### Prerequisites

- [GitLab CLI](https://gitlab.com/gitlab-org/cli) (`glab`), authenticated via `glab auth login`
- Go 1.26 or later (for building from source)

### Install from Source

```bash
# Clone the repository
git clone https://github.com/Omochice/glab-dep.git
cd glab-dep

# Build the extension binary (must be named glab-dep to be a glab extension)
go build -o glab-dep

# Install as a glab extension
glab extension install .
```

Once installed it is invoked as `glab dep`.

## Quick Start

### Interactive Mode (Recommended)

```bash
# Launch interactive TUI across all accessible projects
glab dep

# Or for specific project[s]
glab dep --repo group/app,group/api

# Or for an entire group/subgroup
glab dep --group-path mygroup
```

**In the TUI, you can:**

- Navigate with `↑/↓` or `j/k`
- Select MRs with `space` or `a` (select all)
- Toggle action mode with `m` (Approve → Merge → Approve & Merge)
- Adjust merge settings on-the-fly:
    - `M` - Toggle merge method (squash → merge → rebase)
    - `c` - Toggle CI checks requirement
- Search MRs with `/`
- Open current MR in browser with `o`
- Execute selected actions with `x`
- View help with `?`

### CLI Mode

```bash
# List and group dependency MRs in a single project
glab dep list --repo group/app --group

# Output:
# GROUP              PROJECT   MR     URL
# lodash@4.17.21    app       !123   https://gitlab.com/group/app/-/merge_requests/123
#                   api       !129   https://gitlab.com/group/api/-/merge_requests/129

# View cached groups
glab dep groups

# Approve all MRs in a group (dry-run first)
glab dep approve --group lodash@4.17.21 --dry-run

# Approve for real
glab dep approve --group lodash@4.17.21

# Merge with auto-merge (--require-checks is true by default)
glab dep merge --group lodash@4.17.21 --method squash
```

## Usage

### Scope and authentication

`glab dep` never holds a token. It runs `glab api` / `glab mr` under the hood, so it uses whatever host and credentials `glab` is configured with (run `glab auth login` first). For self-hosted GitLab, run inside a repository of that host or set `GITLAB_HOST`, as documented by `glab`.

The search scope is resolved as follows:

- `--repo` (or `dep.repo`) — explicit `GROUP/PROJECT` paths, comma-separated. Takes precedence.
- `--group-path` — a GitLab group/subgroup full path.
- Neither — searches across all merge requests you can access (`scope=all`).

### Commands

#### Main Command - Interactive TUI (Recommended)

```bash
glab dep [flags]
```

**Flags:**

- `--author` - MR author username (defaults to the Renovate bot, `renovate-bot`)
- `--label` - MR label to filter
- `--reviewer` - Filter MRs by reviewer username
- `--limit` - Max MRs to fetch per project (default: 200)
- `--repo` / `-R` - Target project(s) (`GROUP/PROJECT`), comma-separated
- `--group-path` - Target GitLab group/subgroup full path
- `--mode` - Initial execution mode: `approve`, `merge`, or `approve-and-merge` (default: `approve`)
- `--merge-method` - Initial merge method (default: `squash`)
- `--require-checks` - Merge only when the pipeline succeeds (GitLab auto-merge)

**Examples:**

```bash
# Launch TUI for a single project (defaults to Renovate-authored MRs)
glab dep --repo group/app

# Launch for an entire group with custom initial settings
glab dep --group-path mygroup --merge-method rebase

# Target a custom Renovate bot account
glab dep --group-path mygroup --author my-renovate-bot

# Filter by label
glab dep --repo group/app --label dependencies
```

#### `list` - List dependency MRs

```bash
glab dep list [flags]
```

**Flags:**

- `--author` - MR author username (defaults to the Renovate bot)
- `--label` - MR label to filter
- `--reviewer` - Filter MRs by reviewer username
- `--group` - Group MRs by `package@version` and cache results
- `--json` - Output as JSON
- `--limit` - Max MRs to fetch per project (default: 200)
- `--repo` / `-R` - Target project(s), comma-separated
- `--group-path` - Target GitLab group/subgroup full path

#### `groups` - Show cached groups

```bash
glab dep groups [flags]
```

**Flags:**

- `--json` - Output as JSON

Shows the groups from the last `list --group` command without querying GitLab.

#### `approve` - Bulk approve MRs

```bash
glab dep approve --group GROUP_KEY [flags]
```

**Flags:**

- `--group` - **Required.** Group key (e.g., `lodash@4.17.21`)
- `--dry-run` - Print actions without executing

#### `merge` - Bulk merge MRs

```bash
glab dep merge --group GROUP_KEY [flags]
```

**Flags:**

- `--group` - **Required.** Group key (e.g., `lodash@4.17.21`)
- `--method` - Merge method: `merge`, `squash`, or `rebase` (default: `squash`)
- `--require-checks` - Merge only when the pipeline succeeds via GitLab auto-merge (default: true)
- `--dry-run` - Print actions without executing

With `--require-checks` (the default), each MR is merged through GitLab's native auto-merge: GitLab merges it once its pipeline succeeds, instead of this tool polling and gating the merge. With `--require-checks=false` the MR is merged immediately.

**Examples:**

```bash
# Merge once the pipeline succeeds (recommended)
glab dep merge --group lodash@4.17.21 --method squash

# Merge immediately, regardless of pipeline state
glab dep merge --group lodash@4.17.21 --require-checks=false

# Dry-run merge
glab dep merge --group lodash@4.17.21 --dry-run
```

## Configuration

Save default configuration via `glab config` to avoid passing flags every time:

```bash
# Set default projects
glab config set dep.repo "mygroup/app,mygroup/api,mygroup/web"

# Set the default dependency bot username (defaults to renovate-bot)
glab config set dep.author "my-renovate-bot"

# Set custom MR title patterns (comma-separated regexes with 2 capture groups: package, version)
glab config set dep.patterns "bump\s+([^\s]+)\s+from\s+[^\s]+\s+to\s+v?(\d+(?:\.\d+)?(?:\.\d+)?)"

# View current config
glab config get dep.repo
```

When flags are not provided, `glab dep` uses these defaults.

## Supported MR Title Patterns

The tool automatically parses titles such as:

- `Update dependency <pkg> to vY`
- `chore(deps): update <pkg> to vY`
- `Bump <pkg> from X to Y`

### Custom Patterns

Define your own patterns via `glab config`:

```bash
glab config set dep.patterns "your-pattern-here,another-pattern"
```

**Pattern requirements:**

- Must be valid regex
- Must have exactly 2 capture groups: `(package)` and `(version)`
- Multiple patterns can be comma-separated

**Unknown titles** are grouped as `unknown@unknown` for manual review.

## Cache

Groups are cached at:

```text
${XDG_CACHE_HOME:-$HOME/.cache}/glab-dep/groups.json
```

Cache is overwritten on each `list --group` execution.

## Output Formats

### Human-Readable Tables

```bash
# Flat list
glab dep list
# Output:
# PROJECT                        MR     TITLE
# group/app                     !112   Update dependency lodash to 4.17.21

# Grouped (single table)
glab dep list --group
```

### JSON Output

Use `--json` for machine-readable output:

```bash
# Flat list as JSON array
glab dep list --json

# Grouped as JSON object
glab dep list --group --json
```

## Development

### Build

```bash
go build -o glab-dep
```

### Test

```bash
go test ./...
```

### Run Locally

```bash
./glab-dep --help
```

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure `go test ./...` passes
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Built with:

- [GitLab CLI](https://gitlab.com/gitlab-org/cli) (`glab`)
- [go-gh](https://github.com/cli/go-gh) - table formatter and terminal helpers
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - TUI styling

---

## Made with ❤️ for dependency management automation
