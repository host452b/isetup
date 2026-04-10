# isetup — Project Rules for AI Agents

This file governs how AI agents (Claude Code, Copilot, etc.) should behave when working in this repository.

---

## Philosophy

isetup is a **single-binary CLI tool** for one-click dev environment setup.
Design priority: detect system → resolve install method → execute → report.

**Keep it boring.** No plugin system. No daemon. No config language beyond YAML.
One binary, one config file, one command: `isetup install`.

---

## Language Rules

- **All user-facing output (fmt.Print, log messages, CLI help) must be in English.**
  Chinese comments in code are acceptable, but anything the user sees in terminal must be English.
- **README files:** Maintain `README.md` (English) and `README_zh.md` (Chinese) as sibling files.
  Whenever one changes, the other must be kept in sync.
- **Commit messages:** English. Use conventional commits: `feat:`, `fix:`, `security:`, `docs:`.

---

## Naming Rules

Go identifiers, CLI flags, YAML keys, and file names must be **maximally expressive**.
A reader should understand intent without reading the function body.

- Good: `DetectPkgManagers`, `resolveAptCmd`, `bootstrapPkgs`, `isBootstrapInstalled`
- Bad: `detect`, `resolve`, `pkgs`, `check`

Profile names in `default.yaml` use numbered prefixes (`00-base`, `01-lang-runtimes`) to
enforce alphabetical install order matching the dependency pyramid.

---

## Architecture Rules

- **Single binary.** No external runtime dependencies (no Node, no Python, no Docker).
  `go build -o isetup .` must produce a working binary.
- **Minimal Go dependencies.** Currently: cobra (CLI), yaml.v3 (config), testify (tests).
  Do not add dependencies without strong justification.
- **No network at import time.** Detection, config parsing, and validation are all offline.
  Network calls happen only during tool installation.

---

## Install Method Priority

When resolving how to install a tool on Linux:

1. **Shell with exact OS match** (`shell.linux`) — highest priority
2. **`apt install`** — preferred over apt-get
3. **`apt-get install`** — fallback if apt fails or is absent
4. **`dnf install`** — Fedora/RHEL
5. **`pacman -S`** — Arch
6. **pip** — conda pip preferred, then pip3, then pip
7. **npm** — only if npm is detected

If `apt install` fails at runtime, the executor automatically retries with `apt-get install`.

---

## Bootstrap Rules

On minimal containers (bare `ubuntu:22.04`), basic tools may be missing.
Before running any profile, `bootstrap.go` auto-installs:

- `curl`, `wget`, `ca-certificates`, `gnupg`

These are required by shell-based install scripts (`curl ... | bash`).
Bootstrap uses `apt` first, falls back to `apt-get`.

---

## Security Rules

- **Log file permissions:** `0600` for files, `0700` for directories. Never `0644`.
- **No sensitive env vars in logs.** Do not log PATH, HOME, or any token/secret.
  Only log LANG and system detection info.
- **Use `mktemp` for downloads.** Never hardcode `/tmp/foo.tar.gz` — use `mktemp /tmp/foo-XXXXXX.tar.gz`.
- **Package name validation:** All apt/dnf/pacman/pip/npm package names must pass `safePkgRe`
  (rejects `;`, `|`, `` ` ``, `$()`, `&&`, `>`). Shell commands from YAML are trusted.
- **All URLs must be HTTPS.** No HTTP downloads.
- **Checksum verification:** `install.sh` verifies SHA256. Template scripts should use HTTPS + TLS.
- **sudo handling:** Auto-detect root (UID 0) and omit sudo. Strip sudo from shell commands when root.

---

## Template / default.yaml Rules

- **Pyramid ordering.** Profiles are numbered `00-` through `07-` to enforce install order:
  ```
  00-base → 01-lang-runtimes → 02-git-tools → 03-python-dev
  → 04-ai-tools → 05-shell-enhancements → 06-system-tools → 07-gpu
  ```
- **Explicit `depends_on` for curl users.** Any tool with a `shell:` block that calls `curl` or `wget`
  must declare `depends_on: curl` or `depends_on: wget`.
- **All package managers.** Every apt tool should also have `dnf:`, `pacman:`, `brew:`, `choco:` entries
  where the package exists. Don't leave platforms empty without checking.
- **`when: has_gpu`** is the only supported condition. Don't invent new conditions.
- **Update tool count** in both READMEs when adding/removing tools.

---

## Testing Rules

- Run `go test ./...` before every commit.
- Tests must not require internet, root, or specific OS. Use mock SystemInfo structs.
- Test file naming: `*_test.go` alongside the source file.
- Cover resolver priority, toposort, config validation, and package name safety.

---

## What NOT to Add

- GUI / TUI installer — isetup is CLI-only
- Daemon / service mode — run once, done
- Plugin system — edit default.yaml directly
- Docker/container build integration — isetup runs inside containers, not builds them
- Ansible/Terraform/Chef integration — different layer
- Auto-update mechanism — use `go install` or re-run install.sh
- Interactive prompts during install — all installs are non-interactive (`-y` flags)
