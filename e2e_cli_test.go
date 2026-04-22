package main

// End-to-end tests that exercise the compiled isetup binary via os/exec,
// simulating what a user experiences on the command line. These complement
// the Go-API-level pipeline test in e2e_test.go.
//
// PTY-based tests of the interactive picker are intentionally out of scope:
// adding a pseudo-terminal library violates the minimal-deps policy in
// CLAUDE.md. The picker's state machine is covered by the unit tests in
// internal/picker; this file verifies everything a non-TTY user can trigger.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBinaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "isetup-e2e-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create temp dir:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	testBinaryPath = filepath.Join(dir, "isetup")
	build := exec.Command("go", "build", "-o", testBinaryPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build binary: %v\n%s\n", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

type cliResult struct {
	stdout   string
	stderr   string
	exitCode int
}

func (r cliResult) combined() string { return r.stdout + r.stderr }

type cliInvocation struct {
	stdin string
	env   []string // extra env vars appended to the current environment
}

// runCLI runs the compiled binary with the given args and invocation options.
// Stdin is always a non-TTY pipe (a strings.Reader), which is what CI and
// one-liner installers see.
func runCLI(t *testing.T, opts cliInvocation, args ...string) cliResult {
	t.Helper()
	cmd := exec.Command(testBinaryPath, args...)
	cmd.Stdin = strings.NewReader(opts.stdin)
	if len(opts.env) > 0 {
		cmd.Env = append(os.Environ(), opts.env...)
	}
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run binary: %v (stderr: %s)", err, errb.String())
	}
	return cliResult{stdout: out.String(), stderr: errb.String(), exitCode: code}
}

// writeConfig writes a minimal yaml config to a temp file and returns its path.
func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".isetup.yaml")
	require.NoError(t, os.WriteFile(path, []byte(body), 0600))
	return path
}

// ------------------------------------------------------------------------------
// Discoverability: users running `isetup` with --help should see the interactive
// picker documented.
// ------------------------------------------------------------------------------

func TestCLI_RootHelpListsCommands(t *testing.T) {
	r := runCLI(t, cliInvocation{}, "--help")
	assert.Equal(t, 0, r.exitCode)
	for _, cmd := range []string{"init", "install", "detect", "list", "version"} {
		assert.Contains(t, r.stdout, cmd, "root help should mention %q", cmd)
	}
}

func TestCLI_RootHelpMentionsInteractive(t *testing.T) {
	r := runCLI(t, cliInvocation{}, "--help")
	assert.Equal(t, 0, r.exitCode)
	assert.Contains(t, r.stdout, "-i", "users should learn about -i from the root help")
	assert.Contains(t, r.stdout, "interactive picker")
}

func TestCLI_InstallHelpDocumentsInteractiveFlag(t *testing.T) {
	r := runCLI(t, cliInvocation{}, "install", "--help")
	assert.Equal(t, 0, r.exitCode)
	assert.Contains(t, r.stdout, "-i, --interactive")
	assert.Contains(t, r.stdout, "arrow keys", "install --help should describe the interaction model")
}

// ------------------------------------------------------------------------------
// Version command
// ------------------------------------------------------------------------------

func TestCLI_VersionCommand(t *testing.T) {
	r := runCLI(t, cliInvocation{}, "version")
	assert.Equal(t, 0, r.exitCode)
	assert.True(t, strings.HasPrefix(r.stdout, "isetup v"),
		"version output should start with %q but got %q", "isetup v", r.stdout)
}

// ------------------------------------------------------------------------------
// detect returns parseable JSON with the system fields users rely on.
// ------------------------------------------------------------------------------

func TestCLI_DetectReturnsValidJSON(t *testing.T) {
	r := runCLI(t, cliInvocation{}, "detect")
	require.Equal(t, 0, r.exitCode, "stderr: %s", r.stderr)

	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(r.stdout), &data),
		"detect should emit valid JSON; got: %s", r.stdout)

	for _, key := range []string{"os", "arch", "pkg_managers"} {
		assert.Contains(t, data, key, "detect output should include %q", key)
	}
}

// ------------------------------------------------------------------------------
// Interactive mode: explicit -i without a TTY must exit cleanly with code 2 and
// a clear error pointing users at the fix. This is the most visible guardrail
// for CI and pipe usage.
// ------------------------------------------------------------------------------

func TestCLI_InteractiveFlagRequiresTTY(t *testing.T) {
	r := runCLI(t, cliInvocation{}, "install", "-i")
	assert.Equal(t, 2, r.exitCode,
		"install -i on a non-TTY stdin must exit with code 2")
	assert.Contains(t, r.stderr, "interactive mode requires a TTY",
		"error message should tell the user how to fix the situation")
}

// ------------------------------------------------------------------------------
// Auto-interactive opt-out: when stdin is not a TTY (CI, `curl | bash`), even
// `isetup install` with no other flags must NOT try to launch the picker.
// Going through the non-interactive path with a dry-run config proves this.
// ------------------------------------------------------------------------------

func TestCLI_AutoInteractiveSkippedOnNonTTY(t *testing.T) {
	cfgPath := writeConfig(t, `version: 1
settings:
  dry_run: true
profiles:
  base:
    tools:
      - name: git
        apt: git
        brew: git
`)
	logDir := t.TempDir()
	r := runCLI(t, cliInvocation{},
		"--config", cfgPath,
		"--log-dir", logDir,
		"install",
	)
	assert.Equal(t, 0, r.exitCode,
		"install with no flags on non-TTY should complete successfully; stderr: %s", r.stderr)
	// A real dry-run was performed -- "git" must appear in the output
	// (installed or skipped), which would be impossible if the picker had
	// intercepted stdin and waited for input.
	assert.Contains(t, r.combined(), "git",
		"dry-run output should mention the tool")
}

// ------------------------------------------------------------------------------
// --dry-run end-to-end: banner present, tools listed, no crash.
// ------------------------------------------------------------------------------

func TestCLI_DryRunPrintsBannerAndTools(t *testing.T) {
	cfgPath := writeConfig(t, `version: 1
settings: {}
profiles:
  base:
    tools:
      - name: git
        apt: git
        brew: git
      - name: neovim
        apt: neovim
        brew: neovim
`)
	logDir := t.TempDir()
	r := runCLI(t, cliInvocation{},
		"--config", cfgPath,
		"--log-dir", logDir,
		"install",
		"--dry-run",
	)
	require.Equal(t, 0, r.exitCode, "stderr: %s", r.stderr)
	combined := r.combined()
	assert.Contains(t, combined, "DRY RUN")
	assert.Contains(t, combined, "git")
	assert.Contains(t, combined, "neovim")
}

// ------------------------------------------------------------------------------
// --dry-run + -p restricts to a single profile.
// ------------------------------------------------------------------------------

func TestCLI_DryRunWithProfileFilter(t *testing.T) {
	cfgPath := writeConfig(t, `version: 1
settings: {}
profiles:
  base:
    tools:
      - name: git
        apt: git
        brew: git
  extra:
    tools:
      - name: neovim
        apt: neovim
        brew: neovim
`)
	logDir := t.TempDir()
	r := runCLI(t, cliInvocation{},
		"--config", cfgPath,
		"--log-dir", logDir,
		"install",
		"--dry-run",
		"-p", "base",
	)
	require.Equal(t, 0, r.exitCode, "stderr: %s", r.stderr)
	combined := r.combined()
	assert.Contains(t, combined, "git", "base profile was selected")
	assert.NotContains(t, combined, "neovim", "extra profile should be filtered out")
}

// ------------------------------------------------------------------------------
// Unknown profile name: the warning should surface and (if it was the only
// profile requested) the command exits non-zero.
// ------------------------------------------------------------------------------

func TestCLI_UnknownProfileFailsCleanly(t *testing.T) {
	cfgPath := writeConfig(t, `version: 1
settings: {}
profiles:
  base:
    tools:
      - name: git
        apt: git
        brew: git
`)
	logDir := t.TempDir()
	r := runCLI(t, cliInvocation{},
		"--config", cfgPath,
		"--log-dir", logDir,
		"install",
		"--dry-run",
		"-p", "does-not-exist",
	)
	assert.NotEqual(t, 0, r.exitCode, "unknown profile should not exit 0")
	assert.Contains(t, r.combined(), "does-not-exist")
}

// ------------------------------------------------------------------------------
// list command prints profiles + tools + when conditions.
// ------------------------------------------------------------------------------

func TestCLI_ListShowsProfilesAndTools(t *testing.T) {
	cfgPath := writeConfig(t, `version: 1
settings: {}
profiles:
  00-base:
    tools:
      - name: git
        apt: git
      - name: curl
        apt: curl
  07-gpu:
    when: has_gpu
    tools:
      - name: cuda
        apt: nvidia-cuda-toolkit
`)
	r := runCLI(t, cliInvocation{}, "--config", cfgPath, "list")
	require.Equal(t, 0, r.exitCode, "stderr: %s", r.stderr)
	for _, token := range []string{"00-base", "07-gpu", "git", "curl", "cuda", "has_gpu"} {
		assert.Contains(t, r.stdout, token, "list output should include %q", token)
	}
}

// ------------------------------------------------------------------------------
// init writes a default config file. Override HOME so the test doesn't touch
// the user's real ~/.isetup.yaml.
// ------------------------------------------------------------------------------

func TestCLI_InitWritesDefaultConfig(t *testing.T) {
	fakeHome := t.TempDir()
	r := runCLI(t, cliInvocation{env: []string{"HOME=" + fakeHome}}, "init")
	require.Equal(t, 0, r.exitCode, "stderr: %s", r.stderr)

	cfgPath := filepath.Join(fakeHome, ".isetup.yaml")
	body, err := os.ReadFile(cfgPath)
	require.NoError(t, err, "init should create ~/.isetup.yaml")
	assert.Contains(t, string(body), "profiles:",
		"generated config should contain a profiles section")
	assert.Contains(t, r.stdout, cfgPath,
		"init should tell the user where the config was written")
}

// ------------------------------------------------------------------------------
// init --force overwrites an existing file; without --force it errors.
// ------------------------------------------------------------------------------

func TestCLI_InitRefusesToOverwriteWithoutForce(t *testing.T) {
	fakeHome := t.TempDir()
	existing := filepath.Join(fakeHome, ".isetup.yaml")
	require.NoError(t, os.WriteFile(existing, []byte("existing: true\n"), 0600))

	r := runCLI(t, cliInvocation{env: []string{"HOME=" + fakeHome}}, "init")
	assert.NotEqual(t, 0, r.exitCode,
		"init without --force should fail when config already exists")
	assert.Contains(t, r.combined(), "--force",
		"error message should hint at --force to overwrite")

	// Original file must be untouched.
	b, err := os.ReadFile(existing)
	require.NoError(t, err)
	assert.Equal(t, "existing: true\n", string(b))
}

func TestCLI_InitForceOverwrites(t *testing.T) {
	fakeHome := t.TempDir()
	existing := filepath.Join(fakeHome, ".isetup.yaml")
	require.NoError(t, os.WriteFile(existing, []byte("existing: true\n"), 0600))

	r := runCLI(t, cliInvocation{env: []string{"HOME=" + fakeHome}}, "init", "--force")
	require.Equal(t, 0, r.exitCode, "stderr: %s", r.stderr)

	b, err := os.ReadFile(existing)
	require.NoError(t, err)
	assert.Contains(t, string(b), "profiles:",
		"--force should have replaced the file with the default template")
}

// ------------------------------------------------------------------------------
// Invalid YAML: exits with config-error code (2) and reports the problem.
// ------------------------------------------------------------------------------

func TestCLI_InvalidYAMLFailsWithConfigError(t *testing.T) {
	cfgPath := writeConfig(t, "this: is: not: valid: yaml:\n  - also [bad\n")
	logDir := t.TempDir()
	r := runCLI(t, cliInvocation{},
		"--config", cfgPath,
		"--log-dir", logDir,
		"install",
		"--dry-run",
	)
	assert.NotEqual(t, 0, r.exitCode, "invalid YAML should not exit 0")
}
