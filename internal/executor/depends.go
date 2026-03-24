package executor

import (
	"fmt"

	"github.com/host452b/isetup/internal/config"
)

type ToolEntry struct {
	Tool          config.Tool
	Profile       string
	UnresolvedDep bool
	SkipReason    string
}

func TopoSort(tools []ToolEntry) ([]ToolEntry, error) {
	byName := map[string]int{}
	for i, t := range tools {
		byName[t.Tool.Name] = i
	}

	for i, t := range tools {
		if t.Tool.DependsOn != "" {
			if _, ok := byName[t.Tool.DependsOn]; !ok {
				tools[i].UnresolvedDep = true
			}
		}
	}

	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	for _, t := range tools {
		inDegree[t.Tool.Name] = 0
	}
	for _, t := range tools {
		if t.Tool.DependsOn != "" && !t.UnresolvedDep {
			inDegree[t.Tool.Name]++
			dependents[t.Tool.DependsOn] = append(dependents[t.Tool.DependsOn], t.Tool.Name)
		}
	}

	var queue []string
	for _, t := range tools {
		if inDegree[t.Tool.Name] == 0 {
			queue = append(queue, t.Tool.Name)
		}
	}

	var sorted []ToolEntry
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		idx := byName[name]
		sorted = append(sorted, tools[idx])

		for _, dep := range dependents[name] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(sorted) != len(tools) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return sorted, nil
}
