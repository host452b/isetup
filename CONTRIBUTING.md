# Contributing to isetup

## Adding a New Tool

1. Edit `template/default.yaml` and add the tool to the appropriate profile
2. Include package names for all supported managers: `apt`, `dnf`, `pacman`, `brew`, `choco`
3. If the binary name differs from the tool name, add a case in `internal/executor/installed.go`
4. Run `go test ./...` to verify

## Development

```bash
go build -o isetup .
go test ./... -v
```

Requires Go 1.22+.

## Pull Requests

- One feature/fix per PR
- Include tests for new functionality
- Run `go test ./...` before submitting
- Follow existing code style
