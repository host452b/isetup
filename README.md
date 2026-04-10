# isetup

> Last synced with code: **2026-04-10** · 61 tools · 8 profiles · Go 1.22+

[中文文档](README_zh.md)

Cross-platform CLI tool that detects your OS, hardware, and architecture, then adaptively runs the right install commands to one-click deploy your dev environment.

**CLI-only tools.** isetup is designed for command-line engineers. The default template includes only terminal-based tools — no GUI applications.

New machine? `isetup install`. Done.

## Features

- **Auto-detection** — OS, architecture, distro, GPU, available package managers, shell
- **Multi-platform** — macOS, Linux (Ubuntu/Fedora/Arch), Windows, WSL
- **Profile-based config** — group tools by use case (`00-base`, `01-lang-runtimes`, `02-git-tools`, `03-python-dev`, `04-ai-tools`, `05-shell-enhancements`, `06-system-tools`, `07-gpu`)
- **Adaptive install** — automatically picks `brew`, `apt`, `choco`, `winget`, `dnf`, `pacman`, or custom shell scripts based on what's available
- **Root / Docker aware** — auto-detects UID 0 and omits `sudo` so installs work inside containers without `sudo` installed
- **Template variables** — `{{.Arch}}`, `{{.OS}}`, `{{.Home}}` in shell commands for arch-aware downloads
- **Dependency ordering** — `depends_on` ensures tools install in the right order
- **Conditional profiles** — `when: has_gpu` skips GPU tools on machines without a GPU
- **Skip installed** — auto-detects tools already in PATH, skips them (use `-f` to force reinstall)
- **Rich diagnostics** — full command output, environment snapshot, and timing in `~/.isetup/logs/`
- **Real-time progress** — `[N/Total]` counter with system info header, no silent waiting
- **Dry-run mode** — preview all commands without executing

## Install

**Easiest (if Go is installed):**

```bash
go install github.com/host452b/isetup@latest
```

**One-liner (Linux / macOS):**

```bash
curl -fsSL https://raw.githubusercontent.com/host452b/isetup/main/install.sh | bash
```

**Custom install directory:**

```bash
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/host452b/isetup/main/install.sh | bash
```

**Windows (PowerShell):**

```powershell
$version = (Invoke-RestMethod "https://api.github.com/repos/host452b/isetup/releases/latest").tag_name.TrimStart('v')
$url = "https://github.com/host452b/isetup/releases/download/v$version/isetup_${version}_windows_amd64.zip"
Invoke-WebRequest $url -OutFile isetup.zip
Expand-Archive isetup.zip -DestinationPath .
Move-Item -Force isetup.exe "$env:USERPROFILE\AppData\Local\Microsoft\WindowsApps\"
Remove-Item isetup.zip
```

**From source:**

```bash
git clone https://github.com/host452b/isetup.git && cd isetup && go build -o isetup .
```

**Verify:**

```bash
isetup version
```

## Quick Start

```bash
# Generate default config (optional — isetup works without it)
isetup init

# Edit your config
vim ~/.isetup.yaml

# Preview what will be installed
isetup install --dry-run

# Install everything
isetup install

# Install specific profiles only
isetup install -p 00-base,04-ai-tools
```

## Default Tools

The built-in template installs **61 tools** across 8 profiles:

### lang-runtimes — Language Runtimes & Version Managers

| Tool | Description |
|------|-------------|
| nvm | Node.js version manager — switch between Node versions per project |
| node-lts | Node.js LTS release, installed via nvm |
| typescript | TypeScript compiler (`tsc`) |
| golang | Go programming language |
| rust | Rust toolchain via rustup (rustc, cargo, rustfmt) |
| miniconda | Conda package/env manager — `auto_activate_base` disabled, use `conda activate` manually |
| mise | Polyglot runtime manager (asdf replacement, manages Node/Python/Go/etc. versions) |

### base — Core CLI Essentials

**Editor & Terminal**

| Tool | Description |
|------|-------------|
| git | Distributed version control |
| neovim | Modern terminal editor (Vim fork) |
| tmux | Terminal multiplexer — split panes, detach sessions |
| tmux-ide | Scripted tmux session layouts (npm) |

**Search & Navigation**

| Tool | Description |
|------|-------------|
| fzf | Fuzzy finder — interactive filter for files, history, branches |
| ripgrep | Recursive grep replacement (rg), extremely fast |
| fd | Simpler, faster `find` alternative with sane defaults |
| tree | Visualize directory structure as a tree |

**Modern CLI Replacements (Rust-powered)**

| Tool | Description |
|------|-------------|
| bat | `cat` replacement with syntax highlighting and Git integration |
| eza | `ls` replacement with icons, colors, and Git status |

**Data Processing**

| Tool | Description |
|------|-------------|
| jq | Command-line JSON processor — parse API responses, transform configs |
| yq | YAML/TOML/XML processor — edit CI configs, K8s manifests |

**System Utilities**

| Tool | Description |
|------|-------------|
| htop | Interactive process viewer and system monitor |
| btop | Resource monitor with rich TUI (CPU, memory, disk, network) |
| make | Build automation — many repos ship a Makefile by default |
| curl | URL data transfer (HTTP client, API testing) |
| wget | File downloader with resume support |
| zip | Compression utility |
| unzip | Decompression utility |
| fonts-firacode | Fira Code programming font with ligatures |

### git-tools — Git & CI/CD

| Tool | Description |
|------|-------------|
| gh | GitHub CLI — PRs, issues, actions, repos from terminal |
| glab | GitLab CLI — MRs, pipelines, issues from terminal |
| lazygit | Terminal UI for git — staging, branching, rebasing visually |
| delta | Git diff pager with syntax highlighting and side-by-side view |
| gitlab-runner | GitLab CI runner — run CI jobs locally, manage runners |

### python-dev — Python Ecosystem

| Tool | Description |
|------|-------------|
| uv | Ultra-fast Python package installer (pip replacement) |
| pip-tools | httpie (HTTP client), black (formatter), ruff (linter) |
| pip-build-tools | build, twine, hatchling — Python package publishing |
| huggingface-hub | Hugging Face CLI — download/upload models and datasets |
| pr-analyzers | gitlab-pr-analyzer, github-pr-analyzer, jira-lens |
| playwright | Browser automation for testing (Chromium/Firefox/WebKit) |
| pgcli | PostgreSQL CLI with auto-completion and syntax highlighting |
| ai-ml-libs | chromadb (vector DB), pgvector, langsmith, langfuse (LLM observability) |
| casts-down | Podcast download + Whisper transcription CLI |

### ai-tools — AI & LLM

| Tool | Description |
|------|-------------|
| claude-code | Anthropic Claude Code — AI coding assistant in terminal |
| codex-cli | OpenAI Codex CLI — AI code generation |
| cursor | Cursor AI editor (CLI installer) |
| yoyo | PTY proxy for AI agent auto-approve workflows |
| arxs | Multi-source academic paper search CLI (`go install` preferred, shell fallback) |
| ollama | Run LLMs locally (Llama, Mistral, etc.) |
| llama-cpp | High-performance C++ LLM inference (llama-cli, llama-server) |

### gpu — NVIDIA GPU (conditional: `when: has_gpu`)

| Tool | Description |
|------|-------------|
| cuda-toolkit | NVIDIA CUDA compiler and libraries |
| nvidia-driver | NVIDIA driver v550 |

### shell-enhancements — Shell Productivity

| Tool | Description |
|------|-------------|
| zoxide | Smarter `cd` — learns your most-used directories |
| starship | Cross-shell prompt with git status, language versions, minimal config |
| direnv | Auto-load `.envrc` per directory — manage env vars per project |

### system-tools — Debugging & Networking

| Tool | Description |
|------|-------------|
| lsof | List open files and ports — find what's using port 8080 |
| netcat | TCP/UDP Swiss Army knife — test connections, port scanning |
| tcpdump | Network packet capture and analysis |
| dnsutils | DNS lookup tools: `dig`, `nslookup` |
| strace | Trace system calls — debug process behavior (Linux only) |
| sqlite3 | SQLite database CLI — lightweight DB queries and debugging |
| net-tools | Classic network commands: `ifconfig`, `netstat`, `route` |
| speedtest-cli | Network speed test from the terminal (`speedtest-cli --simple`) |

## Commands

```
isetup init                      Generate default ~/.isetup.yaml
isetup init --force              Overwrite existing config
isetup detect                    Print detected system info as JSON
isetup install                   Install all profiles
isetup install -p 00-base,04-ai-tools  Install specific profiles
isetup install -f                Reinstall even if already installed
isetup install --dry-run         Preview commands without executing
isetup install --timeout 5m     Set per-tool timeout (default 10m)
isetup list                      List all profiles and tools
isetup version                   Print version
```

## Configuration

Config lives at `~/.isetup.yaml` (override with `--config`).

```yaml
version: 1

settings:
  log_level: info
  dry_run: false

profiles:
  00-base:
    tools:
      - name: curl
        apt: curl
        brew: curl

      - name: git
        apt: git
        dnf: git
        pacman: git
        brew: git
        choco: git

  01-lang-runtimes:
    tools:
      - name: nvm
        depends_on: curl
        shell:
          unix: "curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash"

      - name: node-lts
        depends_on: nvm
        shell:
          unix: "source ~/.nvm/nvm.sh && nvm install --lts"

      - name: golang
        brew: go
        depends_on: curl
        shell:
          linux: |
            GO_VERSION=$(curl -fsSL "https://go.dev/VERSION?m=text" | head -1)
            TMPFILE=$(mktemp /tmp/go-XXXXXX.tar.gz)
            curl -fsSL "https://go.dev/dl/${GO_VERSION}.linux-{{.Arch}}.tar.gz" -o "$TMPFILE"
            sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf "$TMPFILE"
            rm -f "$TMPFILE"

  04-ai-tools:
    tools:
      - name: claude-code
        depends_on: node-lts
        shell:
          unix: "curl -fsSL https://claude.ai/install.sh | bash"

      - name: arxs
        depends_on: golang
        shell:
          unix: |
            if command -v go >/dev/null 2>&1; then
              go install github.com/host452b/arxs/v2@latest
            else
              curl -fsSL https://raw.githubusercontent.com/host452b/arxs/main/install.sh | sh
            fi

  07-gpu:
    when: has_gpu
    tools:
      - name: cuda-toolkit
        apt: nvidia-cuda-toolkit
```

> See `template/default.yaml` for the full 61-tool configuration.

### Install Methods

Each tool can declare multiple install methods. isetup picks the best one for the current system:

| Key | Expands to | Platform |
|-----|-----------|----------|
| `apt: X` | `sudo apt install -y X` (fallback: `apt-get`) | Linux (Debian/Ubuntu) — `sudo` omitted when root |
| `dnf: X` | `sudo dnf install -y X` | Linux (Fedora/RHEL) — `sudo` omitted when root |
| `pacman: X` | `sudo pacman -S --noconfirm X` | Linux (Arch) — `sudo` omitted when root |
| `brew: X` | `brew install X` | macOS |
| `choco: X` | `choco install X -y` | Windows (priority) |
| `winget: X` | `winget install --id X -e --accept-source-agreements` | Windows (fallback) |
| `pip: [X, Y]` | `pip3 install X Y` | Any (conda pip preferred) |
| `npm: X` | `npm install -g X` | Any (requires Node.js) |
| `shell:` | Custom commands per OS | Any |

### Shell Priority

```
shell.linux / shell.darwin / shell.windows  (exact OS match)
  → shell.unix                               (linux + darwin fallback)
    → shell (string shorthand)               (unix only)
```

### Template Variables

Shell commands support Go template interpolation:

| Variable | Example value |
|----------|--------------|
| `{{.Arch}}` | `x86_64`, `aarch64`, `arm64` |
| `{{.OS}}` | `linux`, `darwin`, `windows` |
| `{{.Distro}}` | `Ubuntu 22.04.3 LTS` |
| `{{.Home}}` | `/home/user` |

### Conditions

| Condition | Meaning |
|-----------|---------|
| `when: has_gpu` | Skip profile if no GPU detected |

## Logging

Logs are written to `~/.isetup/logs/` (override with `--log-dir`).

Each run produces two files:

- `isetup-<timestamp>.env.json` — environment snapshot (OS, arch, GPU, pkg managers — no sensitive env vars)
- `isetup-<timestamp>.log` — per-tool install record with command, stdout, stderr, exit code, duration

Example terminal output:

```
Detecting system...
OS: linux | Arch: amd64 | Shell: /bin/bash
Package managers: apt, pip3, npm
GPU: NVIDIA H200 NVL

[1/61] Installing nvm (shell: curl -o- https://nvm.sh/install.sh | bash)...
[1/61] nvm                  PASS    (shell ) 0.7s
[2/61] Installing node-lts (shell: source ~/.nvm/nvm.sh && nvm install --lts)...
[2/61] node-lts             PASS    (shell ) 0.5s
[3/61] git                  SKIP    already installed
[4/61] Installing glab (shell: curl ... | sudo bash)...
[4/61] glab                 PASS    (shell ) 2.1s
[5/61] cuda-toolkit         FAILED  (apt   ) 1.1s
       E: Unable to locate package nvidia-cuda-toolkit

─────────────────────────────
Installed: 50 | Failed: 1 | Skipped: 6
Log: ~/.isetup/logs/isetup-2026-03-25T00-04-21.log
```

Output is color-coded: green for PASS, red for FAILED, yellow for SKIP. First line of stderr is shown inline for failed tools.

## System Detection

`isetup detect` outputs full system info as JSON:

```json
{
  "os": "darwin",
  "arch": "arm64",
  "arch_label": "arm64",
  "distro": "macOS 15.3.2",
  "kernel": "24.3.0",
  "wsl": false,
  "is_root": false,
  "shell": "/bin/zsh",
  "gpu": {
    "detected": true,
    "model": "Apple M3 Pro"
  },
  "pkg_managers": ["brew", "pip3", "npm"]
}
```

## Project Structure

```
isetup/
├── main.go                  # Entry point
├── embed.go                 # Embeds default template
├── cmd/                     # CLI commands (cobra)
├── internal/
│   ├── config/              # YAML parsing + validation
│   ├── detector/            # OS/GPU/shell/pkg manager detection
│   ├── executor/            # Install engine (bootstrap, resolver, runner, topo sort)
│   └── logger/              # Structured logging
└── template/
    └── default.yaml         # Default config template
```

## Install Flow

What happens when you run `isetup install`:

```
┌─────────────────────────────────────────────────────────────┐
│  isetup install                                              │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  1. DETECT SYSTEM                                            │
│     OS, arch, distro, GPU, shell, package managers          │
│     → SystemInfo struct                                      │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  2. LOAD CONFIG                                              │
│     ~/.isetup.yaml → parse YAML → validate                  │
│     (if no config: use embedded default.yaml)               │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  3. BOOTSTRAP (minimal containers only)                      │
│     Missing curl/wget/ca-certificates/gnupg?                │
│     → apt update && apt install -y ...                       │
│     → fallback: apt-get if apt fails                        │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  4. COLLECT & SORT TOOLS                                     │
│     Profiles sorted alphabetically (00-base → 07-gpu)       │
│     Topological sort by depends_on                          │
│     → ordered tool list                                      │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  5. FOR EACH TOOL:                                           │
│                                                              │
│     ┌── already in PATH?  → SKIP                            │
│     ├── depends_on failed? → SKIP                           │
│     ├── when: condition not met? → SKIP                     │
│     └── resolve install method:                             │
│         shell.linux > apt > dnf > pacman > pip > npm        │
│                                                              │
│     Execute command (timeout: 10m default)                   │
│     If apt fails → retry with apt-get                       │
│     If root → strip sudo from commands                      │
│                                                              │
│     Log: stdout, stderr, exit code, duration                │
│     Report: PASS / FAIL / SKIP                              │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  6. SUMMARY                                                  │
│     Installed: N | Failed: N | Skipped: N                   │
│     Log: ~/.isetup/logs/isetup-<timestamp>.log              │
└─────────────────────────────────────────────────────────────┘
```

### Install Method Resolution (per tool)

```
Has shell.linux / shell.darwin / shell.windows?
  → YES: use it (highest priority — exact OS match)
  → NO: check shell.unix fallback
         → YES: use it (linux + darwin)
         → NO: try package managers:
                apt → apt-get (fallback) → dnf → pacman → brew → choco → winget
                → pip (conda pip > pip3 > pip)
                → npm
                → no method found: SKIP
```

## Troubleshooting

**Tool install hangs:**
Use `--timeout 2m` to set a shorter per-tool timeout. Check `~/.isetup/logs/` for the full command output.

**Minimal container (no curl/wget):**
isetup auto-detects missing prerequisites (curl, wget, ca-certificates, gnupg) and installs them via apt/apt-get before running any profiles. Works out-of-the-box in bare `ubuntu:22.04` Docker containers.

**Permission denied:**
isetup auto-detects root and omits `sudo`. If you're not root and `sudo` fails, ensure your user has sudo privileges.

**Profile not found:**
Check available profiles with `isetup list`. Profile names are case-sensitive.

**Already installed tools:**
isetup skips tools found in PATH. Use `-f` to force reinstall.

**Custom config location:**
Use `--config /path/to/config.yaml` to point to a non-default config file.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All tools installed or skipped successfully |
| 1 | One or more tools failed to install |
| 2 | Configuration error (invalid YAML, validation failure) |

## Build

```bash
go build -o isetup .
```

Requires Go 1.22+.

## License

MIT
