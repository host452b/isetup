package executor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/host452b/isetup/internal/config"
	"github.com/host452b/isetup/internal/detector"
	"github.com/host452b/isetup/internal/logger"
)

// ProgressCallback is called before and after each tool install.
type ProgressCallback func(event ProgressEvent)

// ProgressEvent describes what's happening during installation.
type ProgressEvent struct {
	Index   int    // 0-based tool index
	Total   int    // total tool count
	Name    string // tool name
	Profile string // profile name
	Phase   string // "start", "done"
	Method  string // install method (available on "done")
	Command string // resolved command (available on "start" for non-skip)
	Result  *logger.ToolResult // available on "done"
}

// Execute runs the full install pipeline. Returns results and an error if topology is invalid.
func Execute(ctx context.Context, cfg *config.Config, info *detector.SystemInfo, lg *logger.Logger, profiles []string, toolFilter []string, onProgress ProgressCallback) ([]logger.ToolResult, error) {
	// Bootstrap: ensure minimal prerequisites exist (curl, wget, ca-certificates)
	if !cfg.Settings.DryRun {
		Bootstrap(ctx, info, lg)
	}

	entries := collectTools(cfg, info, profiles, toolFilter)

	sorted, err := TopoSort(entries)
	if err != nil {
		return nil, fmt.Errorf("dependency error: %w", err)
	}

	total := len(sorted)
	var results []logger.ToolResult
	failed := map[string]bool{}

	for i, entry := range sorted {
		result := logger.ToolResult{
			Name:    entry.Tool.Name,
			Profile: entry.Profile,
		}

		// Check if interrupted
		if ctx.Err() != nil {
			result.Status = logger.StatusSkipped
			result.SkipReason = "interrupted"
			_ = lg.WriteToolResult(result)
			results = append(results, result)
			notify(onProgress, i, total, entry, "done", "", "", &result)
			continue
		}

		// Check depends_on
		if entry.Tool.DependsOn != "" {
			if failed[entry.Tool.DependsOn] {
				result.Status = logger.StatusSkipped
				result.SkipReason = fmt.Sprintf("dependency failed: %s", entry.Tool.DependsOn)
				failed[entry.Tool.Name] = true
				_ = lg.WriteToolResult(result)
				results = append(results, result)
				notify(onProgress, i, total, entry, "done", "", "", &result)
				continue
			}
			// Unresolved dep = not in selected profiles. Check if already on system.
			if entry.UnresolvedDep {
				depTool := config.Tool{Name: entry.Tool.DependsOn}
				if !IsInstalled(depTool) {
					result.Status = logger.StatusSkipped
					result.SkipReason = fmt.Sprintf("dependency not available: %s (install profile containing it first)", entry.Tool.DependsOn)
					failed[entry.Tool.Name] = true
					_ = lg.WriteToolResult(result)
					results = append(results, result)
					notify(onProgress, i, total, entry, "done", "", "", &result)
					continue
				}
				// dep is already installed on system — proceed
			}
		}

		// Check when condition (carried on entry from collectTools)
		if entry.SkipReason != "" {
			result.Status = logger.StatusSkipped
			result.SkipReason = entry.SkipReason
			_ = lg.WriteToolResult(result)
			results = append(results, result)
			notify(onProgress, i, total, entry, "done", "", "", &result)
			continue
		}

		// Check if already installed (skip unless --force)
		if !cfg.Settings.Force && IsInstalled(entry.Tool) {
			result.Status = logger.StatusSkipped
			result.SkipReason = "already installed"
			_ = lg.WriteToolResult(result)
			results = append(results, result)
			notify(onProgress, i, total, entry, "done", "", "", &result)
			continue
		}

		// Resolve install method
		method, cmd := Resolve(entry.Tool, info)
		if method == "" {
			result.Status = logger.StatusSkipped
			result.SkipReason = fmt.Sprintf("no install method for %s on %s", entry.Tool.Name, info.OS)
			failed[entry.Tool.Name] = true
			_ = lg.WriteToolResult(result)
			results = append(results, result)
			notify(onProgress, i, total, entry, "done", "", "", &result)
			continue
		}

		// Interpolate template variables in shell commands
		if method == "shell" {
			interpolated, err := Interpolate(cmd, info)
			if err != nil {
				result.Status = logger.StatusFailed
				result.SkipReason = fmt.Sprintf("template error: %v", err)
				failed[entry.Tool.Name] = true
				_ = lg.WriteToolResult(result)
				results = append(results, result)
				notify(onProgress, i, total, entry, "done", method, cmd, &result)
				continue
			}
			cmd = interpolated
		}

		// Strip sudo when running as root (e.g. inside Docker containers)
		if info.IsRoot && method == "shell" {
			cmd = StripSudo(cmd)
		}

		result.Method = method
		result.Command = cmd

		// Dry run
		if cfg.Settings.DryRun {
			notify(onProgress, i, total, entry, "start", method, cmd, nil)
			result.Status = logger.StatusSuccess
			_ = lg.WriteToolResult(result)
			results = append(results, result)
			notify(onProgress, i, total, entry, "done", method, cmd, &result)
			continue
		}

		// Notify start
		notify(onProgress, i, total, entry, "start", method, cmd, nil)

		// Execute
		timeout := 10 * time.Minute
		if cfg.Settings.Timeout > 0 {
			timeout = cfg.Settings.Timeout
		}
		toolCtx, toolCancel := context.WithTimeout(ctx, timeout)
		runResult := Run(toolCtx, cmd, info.Shell)
		toolCancel()

		// apt → apt-get fallback: if "apt install" fails, retry with "apt-get install"
		if runResult.ExitCode != 0 && method == "apt" && strings.Contains(cmd, "apt install") {
			fallbackCmd := strings.Replace(cmd, "apt install", "apt-get install", 1)
			toolCtx2, toolCancel2 := context.WithTimeout(ctx, timeout)
			runResult = Run(toolCtx2, fallbackCmd, info.Shell)
			toolCancel2()
			if runResult.ExitCode == 0 {
				cmd = fallbackCmd
			}
		}

		result.ExitCode = runResult.ExitCode
		result.Duration = runResult.Duration
		result.Stdout = runResult.Stdout
		result.Stderr = runResult.Stderr
		result.Command = cmd

		if runResult.ExitCode == 0 {
			result.Status = logger.StatusSuccess
		} else {
			result.Status = logger.StatusFailed
			failed[entry.Tool.Name] = true
		}

		_ = lg.WriteToolResult(result)
		results = append(results, result)
		notify(onProgress, i, total, entry, "done", method, cmd, &result)
	}

	return results, nil
}

func notify(cb ProgressCallback, index, total int, entry ToolEntry, phase, method, command string, result *logger.ToolResult) {
	if cb == nil {
		return
	}
	cb(ProgressEvent{
		Index:   index,
		Total:   total,
		Name:    entry.Tool.Name,
		Profile: entry.Profile,
		Phase:   phase,
		Method:  method,
		Command: command,
		Result:  result,
	})
}

func collectTools(cfg *config.Config, info *detector.SystemInfo, profileFilter, toolFilter []string) []ToolEntry {
	selected := cfg.Profiles
	if profileFilter != nil {
		selected = make(map[string]config.Profile)
		for _, name := range profileFilter {
			if p, ok := cfg.Profiles[name]; ok {
				selected[name] = p
			}
		}
	}

	toolSet := map[string]bool(nil)
	if toolFilter != nil {
		toolSet = make(map[string]bool, len(toolFilter))
		for _, n := range toolFilter {
			toolSet[n] = true
		}
	}

	names := make([]string, 0, len(selected))
	for name := range selected {
		names = append(names, name)
	}
	sort.Strings(names)

	var entries []ToolEntry
	for _, profName := range names {
		prof := selected[profName]
		skipReason := ""
		if prof.When != "" && !evaluateCondition(prof.When, info) {
			skipReason = fmt.Sprintf("condition not met: %s", prof.When)
		}

		for _, tool := range prof.Tools {
			if toolSet != nil && !toolSet[tool.Name] {
				continue
			}
			entries = append(entries, ToolEntry{
				Tool:       tool,
				Profile:    profName,
				SkipReason: skipReason,
			})
		}
	}
	return entries
}

func evaluateCondition(when string, info *detector.SystemInfo) bool {
	switch when {
	case "has_gpu":
		return info.GPU.Detected
	default:
		return false
	}
}
