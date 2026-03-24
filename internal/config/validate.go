package config

import "fmt"

var validWhenConditions = map[string]bool{
	"has_gpu": true,
}

func Validate(cfg *Config) ([]error, []string) {
	var errs []error
	var warns []string

	if cfg.Version == 0 {
		errs = append(errs, fmt.Errorf("missing version field"))
	} else if cfg.Version != 1 {
		errs = append(errs, fmt.Errorf("unsupported config version: %d", cfg.Version))
	}

	toolNames := map[string]string{}
	allTools := []Tool{}

	for profName, prof := range cfg.Profiles {
		if len(prof.Tools) == 0 {
			warns = append(warns, fmt.Sprintf("profile %q has empty tools list", profName))
		}

		if prof.When != "" {
			if !validWhenConditions[prof.When] {
				errs = append(errs, fmt.Errorf("unknown when condition %q in profile %q", prof.When, profName))
			}
		}

		for _, tool := range prof.Tools {
			if tool.Name == "" {
				errs = append(errs, fmt.Errorf("tool in profile %q has missing name", profName))
				continue
			}

			if existingProf, ok := toolNames[tool.Name]; ok {
				errs = append(errs, fmt.Errorf("duplicate tool name %q in profiles %q and %q", tool.Name, existingProf, profName))
			} else {
				toolNames[tool.Name] = profName
			}

			allTools = append(allTools, tool)
		}
	}

	for _, tool := range allTools {
		if tool.DependsOn != "" {
			if _, ok := toolNames[tool.DependsOn]; !ok {
				warns = append(warns, fmt.Sprintf("tool %q depends_on %q which does not exist", tool.Name, tool.DependsOn))
			}
		}
	}

	if err := checkCircular(allTools); err != nil {
		errs = append(errs, err)
	}

	return errs, warns
}

func checkCircular(tools []Tool) error {
	deps := map[string]string{}
	for _, t := range tools {
		if t.DependsOn != "" {
			deps[t.Name] = t.DependsOn
		}
	}

	for _, t := range tools {
		visited := map[string]bool{}
		cur := t.Name
		for {
			if visited[cur] {
				return fmt.Errorf("circular dependency detected involving %q", cur)
			}
			visited[cur] = true
			next, ok := deps[cur]
			if !ok {
				break
			}
			cur = next
		}
	}
	return nil
}
