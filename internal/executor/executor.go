package executor

import (
	"fmt"
	"sort"

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
func Execute(cfg *config.Config, info *detector.SystemInfo, lg *logger.Logger, profiles []string, onProgress ProgressCallback) ([]logger.ToolResult, error) {
	entries := collectTools(cfg, info, profiles)

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

		// Check depends_on
		if entry.Tool.DependsOn != "" {
			if entry.UnresolvedDep || failed[entry.Tool.DependsOn] {
				result.Status = logger.StatusSkipped
				result.SkipReason = fmt.Sprintf("dependency failed: %s", entry.Tool.DependsOn)
				failed[entry.Tool.Name] = true
				_ = lg.WriteToolResult(result)
				results = append(results, result)
				notify(onProgress, i, total, entry, "done", "", "", &result)
				continue
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
		runResult := Run(cmd, info.Shell)
		result.ExitCode = runResult.ExitCode
		result.Duration = runResult.Duration
		result.Stdout = runResult.Stdout
		result.Stderr = runResult.Stderr

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

func collectTools(cfg *config.Config, info *detector.SystemInfo, profileFilter []string) []ToolEntry {
	selected := cfg.Profiles
	if profileFilter != nil {
		selected = make(map[string]config.Profile)
		for _, name := range profileFilter {
			if p, ok := cfg.Profiles[name]; ok {
				selected[name] = p
			}
		}
	}

	// Deterministic profile order
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
