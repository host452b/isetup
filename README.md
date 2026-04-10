# isetup

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

The built-in template installs **58 tools** across 8 profiles:

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

### ai-tools — AI & LLM

| Tool | Description |
|------|-------------|
| claude-code | Anthropic Claude Code — AI coding assistant in terminal |
| codex-cli | OpenAI Codex CLI — AI code generation |
| cursor | Cursor AI editor (CLI installer) |
| yoyo | PTY proxy for AI agent auto-approve workflows |
| arxs | Multi-source academic paper search CLI |
| ollama | Run LLMs locally (Llama, Mistral, etc.) |

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
  base:
    tools:
      - name: git
        apt: git
        dnf: git
        pacman: git
        brew: git
        choco: git

      - name: neovim
        apt: neovim
        brew: neovim
        choco: neovim

  lang-runtimes:
    tools:
      - name: nvm
        shell:
          unix: |
            curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash

      - name: node-lts
        depends_on: nvm
        shell:
          unix: "source ~/.nvm/nvm.sh && nvm install --lts"

      - name: golang
        brew: go
        shell:
          linux: |
            GO_VERSION=$(curl -fsSL "https://go.dev/VERSION?m=text" | head -1)
            curl -fsSL "https://go.dev/dl/${GO_VERSION}.linux-{{.Arch}}.tar.gz" -o /tmp/go.tar.gz
            sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tar.gz

      - name: rust
        shell:
          unix: "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y"

      - name: miniconda
        shell:
          linux: |
            curl -fsSL https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-{{.Arch}}.sh -o /tmp/miniconda.sh
            bash /tmp/miniconda.sh -b -p $HOME/miniconda3
            $HOME/miniconda3/bin/conda config --set auto_activate_base false
            $HOME/miniconda3/bin/conda init
          darwin: |
            curl -fsSL https://repo.anaconda.com/miniconda/Miniconda3-latest-MacOSX-{{.Arch}}.sh -o /tmp/miniconda.sh
            bash /tmp/miniconda.sh -b -p $HOME/miniconda3
            $HOME/miniconda3/bin/conda config --set auto_activate_base false
            $HOME/miniconda3/bin/conda init

      - name: pip-tools
        depends_on: miniconda
        pip:
          - httpie
          - black
          - ruff

  git-tools:
    tools:
      - name: glab
        brew: glab
        shell:
          linux: |
            curl -fsSL "https://packages.gitlab.com/install/repositories/gitlab/glab/script.deb.sh" | sudo bash
            sudo apt-get install -y glab

      - name: gh
        brew: gh
        shell:
          linux: |
            # Official GitHub CLI repo
            wget -qO- https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null
            echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
            sudo apt update && sudo apt install gh -y

  ai-tools:
    tools:
      - name: claude-code
        depends_on: node-lts
        shell:
          unix: "curl -fsSL https://claude.ai/install.sh | bash"
          windows: "irm https://claude.ai/install.ps1 | iex"

      - name: codex-cli
        depends_on: node-lts
        npm: "@openai/codex"

      - name: cursor
        shell:
          unix: "curl -fsS https://cursor.com/install | bash"

      - name: yoyo
        shell:
          unix: "curl -fsSL https://github.com/host452b/yoyo/releases/latest/download/install.sh | sh"

      - name: arxs
        shell:
          unix: "curl -fsSL https://raw.githubusercontent.com/host452b/arxs/main/install.sh | sh"
          windows: "irm https://raw.githubusercontent.com/host452b/arxs/main/install.ps1 | iex"

  gpu:
    when: has_gpu
    tools:
      - name: cuda-toolkit
        apt: nvidia-cuda-toolkit
      - name: nvidia-driver
        apt: nvidia-driver-550
```

### Install Methods

Each tool can declare multiple install methods. isetup picks the best one for the current system:

| Key | Expands to | Platform |
|-----|-----------|----------|
| `apt: X` | `sudo apt-get install -y X` | Linux (Debian/Ubuntu) — `sudo` omitted when root |
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

- `isetup-<timestamp>.env.json` — full environment snapshot (OS, arch, GPU, PATH, etc.)
- `isetup-<timestamp>.log` — per-tool install record with command, stdout, stderr, exit code, duration

Example terminal output:

```
Detecting system...
OS: linux | Arch: amd64 | Shell: /bin/bash
Package managers: apt, pip3, npm
GPU: NVIDIA H200 NVL

[1/57] Installing nvm (shell: curl -o- https://nvm.sh/install.sh | bash)...
[1/57] nvm                  PASS    (shell ) 0.7s
[2/57] Installing node-lts (shell: source ~/.nvm/nvm.sh && nvm install --lts)...
[2/57] node-lts             PASS    (shell ) 0.5s
[3/57] git                  SKIP    already installed
[4/57] Installing glab (shell: curl ... | sudo bash)...
[4/57] glab                 PASS    (shell ) 2.1s
[5/57] cuda-toolkit         FAILED  (apt   ) 1.1s
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
│   ├── executor/            # Install engine (resolver, runner, topo sort)
│   └── logger/              # Structured logging
└── template/
    └── default.yaml         # Default config template
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
