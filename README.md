# isetup

[中文文档](README_zh.md)

Cross-platform CLI tool that detects your OS, hardware, and architecture, then adaptively runs the right install commands to one-click deploy your dev environment.

**CLI-only tools.** isetup is designed for command-line engineers. The default template includes only terminal-based tools — no GUI applications.

New machine? `isetup install`. Done.

## Features

- **Auto-detection** — OS, architecture, distro, GPU, available package managers, shell
- **Multi-platform** — macOS, Linux (Ubuntu/Fedora/Arch), Windows, WSL
- **Profile-based config** — group tools by use case (`base`, `node-dev`, `python-dev`, `ai-tools`, `gpu`)
- **Adaptive install** — automatically picks `brew`, `apt`, `choco`, `winget`, `dnf`, `pacman`, or custom shell scripts based on what's available
- **Template variables** — `{{.Arch}}`, `{{.OS}}`, `{{.Home}}` in shell commands for arch-aware downloads
- **Dependency ordering** — `depends_on` ensures tools install in the right order
- **Conditional profiles** — `when: has_gpu` skips GPU tools on machines without a GPU
- **Rich diagnostics** — full command output, environment snapshot, and timing in `~/.isetup/logs/`
- **Dry-run mode** — preview all commands without executing

## Install

**Linux / macOS (recommended):**

```bash
curl -fsSL https://raw.githubusercontent.com/host452b/isetup/main/install.sh | bash
```

Or install to a custom directory:

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

**Go install:**

```bash
go install github.com/isetup-dev/isetup@latest
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
isetup install -p base,ai-tools
```

## Commands

```
isetup init                      Generate default ~/.isetup.yaml
isetup init --force              Overwrite existing config
isetup detect                    Print detected system info as JSON
isetup install                   Install all profiles
isetup install -p base,node-dev  Install specific profiles
isetup install --dry-run         Preview commands without executing
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
        brew: git
        choco: git

      - name: neovim
        apt: neovim
        brew: neovim
        choco: neovim

  node-dev:
    tools:
      - name: nvm
        shell:
          unix: |
            curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
          windows: |
            irm https://github.com/coreybutler/nvm-windows/releases/download/1.1.12/nvm-setup.exe -OutFile nvm-setup.exe

      - name: node-lts
        depends_on: nvm
        shell:
          unix: "source ~/.nvm/nvm.sh && nvm install --lts"
          windows: "nvm install lts && nvm use lts"

  python-dev:
    tools:
      - name: miniconda
        shell:
          linux: |
            curl -fsSL https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-{{.Arch}}.sh -o /tmp/miniconda.sh
            bash /tmp/miniconda.sh -b -p $HOME/miniconda3
            $HOME/miniconda3/bin/conda init
          darwin: |
            curl -fsSL https://repo.anaconda.com/miniconda/Miniconda3-latest-MacOSX-{{.Arch}}.sh -o /tmp/miniconda.sh
            bash /tmp/miniconda.sh -b -p $HOME/miniconda3
            $HOME/miniconda3/bin/conda init

      - name: pip-tools
        depends_on: miniconda
        pip:
          - httpie
          - black
          - ruff

  ai-tools:
    tools:
      - name: claude-code
        shell:
          unix: "curl -fsSL https://claude.ai/install.sh | bash"
          windows: "irm https://claude.ai/install.ps1 | iex"

      - name: codex-cli
        npm: "@openai/codex"

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
| `apt: X` | `sudo apt-get install -y X` | Linux (Debian/Ubuntu) |
| `dnf: X` | `sudo dnf install -y X` | Linux (Fedora/RHEL) |
| `pacman: X` | `sudo pacman -S --noconfirm X` | Linux (Arch) |
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
git                  PASS    (brew  ) 0.8s
neovim               PASS    (brew  ) 3.2s
nvm                  PASS    (shell ) 2.1s
cuda-toolkit         FAILED  (apt   ) 1.1s  → see log
─────────────────────────────
Installed: 3 | Failed: 1 | Skipped: 0
Log: ~/.isetup/logs/isetup-2026-03-24T20-57-30.log
```

Output is color-coded: green for PASS, red for FAILED, yellow for SKIP.

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

## Build

```bash
go build -o isetup .
```

Requires Go 1.22+.

## License

MIT
