package picker

import (
	"sort"

	"github.com/host452b/isetup/internal/config"
)

// ResolveDeps computes the transitive closure of `selected` under the
// `DependsOn` relation defined by `all`. Dependencies that name tools not
// present in `all` are dropped from the closure (the executor's runtime path
// handles "already on system" detection for those).
//
// Returns:
//
//	closure — all tools that must be considered, sorted by name.
//	added   — tools added by dependency resolution (closure − selected), sorted.
func ResolveDeps(selected []string, all []config.Tool) (closure, added []string) {
	byName := make(map[string]config.Tool, len(all))
	for _, t := range all {
		byName[t.Name] = t
	}

	selectedSet := make(map[string]bool, len(selected))
	for _, s := range selected {
		selectedSet[s] = true
	}

	inClosure := make(map[string]bool)
	queue := append([]string(nil), selected...)
	for len(queue) > 0 {
		t := queue[0]
		queue = queue[1:]
		if inClosure[t] {
			continue
		}
		// Only add to closure if it exists in byName or was explicitly selected
		if _, ok := byName[t]; !ok && !selectedSet[t] {
			continue
		}
		inClosure[t] = true
		if tool, ok := byName[t]; ok && tool.DependsOn != "" {
			queue = append(queue, tool.DependsOn)
		}
	}

	for name := range inClosure {
		closure = append(closure, name)
		if !selectedSet[name] {
			added = append(added, name)
		}
	}
	sort.Strings(closure)
	sort.Strings(added)
	return
}
