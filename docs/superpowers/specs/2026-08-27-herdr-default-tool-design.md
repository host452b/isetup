# Add Herdr to the Default AI Tools Profile

- **Status**: Approved
- **Date**: 2026-08-27
- **Target release**: unreleased / next release
- **Related**: `template/default.yaml`, `README.md`, `README_zh.md`

## Motivation

Herdr is a terminal runtime for coding agents. It belongs in isetup's default
`04-ai-tools` profile alongside Claude Code, Codex CLI, and the other agent
tools.

## Design

Add one tool named `herdr` to `04-ai-tools` with `depends_on: curl`.
Linux and macOS use Herdr's official Unix installer:

```sh
curl -fsSL https://herdr.dev/install.sh | sh
```

Windows uses Herdr's official PowerShell installer:

```powershell
powershell -ExecutionPolicy Bypass -c "irm https://herdr.dev/install.ps1 | iex"
```

The Unix installer is used on both Linux and macOS to keep the default config
aligned with the requested canonical command. Homebrew and mise are documented
upstream alternatives but are not needed in this tool entry.

## Documentation

Update both README files together:

- Add Herdr to the AI tools table as a persistent terminal runtime for coding
  agents.
- Change the built-in tool count from 62 to 63.
- Change the interactive picker example for `04-ai-tools` from 7 to 8 tools.
- Keep any other historical example counters unchanged unless they describe
  the current built-in template total.

## Validation

Follow test-driven development:

1. Add a contract test that loads the embedded default template and asserts
   that `04-ai-tools` contains `herdr`, depends on `curl`, and has the exact
   official Unix and Windows commands.
2. Run that test and confirm it fails because Herdr is absent.
3. Add the minimal YAML entry and run the focused test again.
4. Synchronize both README files, then run `go test ./...`.

No installer is executed during testing; all validation remains offline as
required by the repository rules.

## Sources

- https://github.com/herdrdev/herdr#install
- https://herdr.dev/install.sh
