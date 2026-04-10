# Changelog

All notable changes to isetup are documented here.

## [1.1.0] - 2026-04-10

### Added
- **Bootstrap for minimal containers**: auto-installs curl, wget, ca-certificates, gnupg before any profile runs — bare `ubuntu:22.04` now works out-of-the-box
- **One-liner install**: `curl ... | bash` now auto-runs `isetup install` after downloading (set `ISETUP_NO_AUTO_INSTALL=1` to skip)
- **Pyramid dependency ordering**: profiles numbered `00-` through `07-` to enforce correct install order
- **Explicit `depends_on` for curl/wget**: all shell-based tools now declare their download dependency
- **Failure diagnostics**: failed tools print first 3 lines of stderr, a copy-pasteable retry command, and a summary with "paste to AI for diagnosis" tip
- **Version pinning**: nvm 0.40.1, Node 22.15.0, Go 1.24.2, Rust 1.86.0
- **New tools**: glow, net-tools, llama-cpp, casts-down, speedtest-cli (62 total)
- **arxs updated**: prefers `go install github.com/host452b/arxs/v2@latest`, falls back to curl
- **Install flow diagram** in README
- **CLAUDE.md**: project rules for AI agents

### Changed
- `apt install` is now preferred over `apt-get install` (runtime fallback: if apt fails, retries with apt-get)
- Detect both `apt` and `apt-get` as separate package managers
- README Configuration example updated to match actual `default.yaml` (numbered profiles, mktemp, depends_on)
- README freshness timestamp added at top

### Security
- Log file permissions: `0644` → `0600`, directory `0755` → `0700` (owner-only)
- Removed PATH and HOME from env.json logs (potential info leak)
- All `/tmp/` hardcoded paths replaced with `mktemp` (race condition fix)

## [Unreleased]

### Added
- Command execution timeout with `--timeout` flag (default 10m per tool)
- Graceful Ctrl+C with interrupt summary
- Profile name validation with typo suggestions
- Differentiated exit codes (0=ok, 1=partial-fail, 2=config-error)
- Actionable error messages with fix suggestions
- Usage examples on all subcommands
- Log path shown at start of install
- dnf/pacman package support for Fedora and Arch Linux
- Version injection via ldflags
- New tools: bat, eza, fd, jq, yq, tree, htop, btop, fonts-firacode, make, curl, wget, zip, unzip, lazygit, delta, mise, huggingface-hub, playwright, chromadb, pgvector, langsmith, langfuse, ollama, zoxide, starship, direnv, lsof, netcat, tcpdump, dnsutils, strace, sqlite3, pgcli, gitlab-runner
- Troubleshooting and exit codes documentation in README

### Fixed
- Command injection vulnerability in package name handling
- Log files could grow unbounded (now truncated at 10KB per field)

### Security
- Package names validated against safe-character regex before shell interpolation

## [0.3.0] - 2026-03-25

### Added
- Root/Docker support: omit `sudo` when running as UID 0

## [0.2.0] - 2026-03-20

### Added
- Skip already-installed tools (use `-f` to force)
- Language runtime profile
- Module path fix

## [0.1.0] - 2026-03-15

### Added
- Initial release: cross-platform dev environment setup
- Profile-based YAML config
- Auto-detection of OS, arch, GPU, package managers
- Dependency ordering with topological sort
- Dry-run mode
- Structured logging
