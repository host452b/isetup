# Changelog

All notable changes to isetup are documented here.

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
