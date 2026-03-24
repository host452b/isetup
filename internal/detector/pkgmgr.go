package detector

import "os/exec"

var knownPkgManagers = []string{
	"apt", "dnf", "pacman", "brew", "choco", "winget", "pip3", "pip", "npm",
}

func DetectPkgManagers() []string {
	var found []string
	for _, pm := range knownPkgManagers {
		if isInPath(pm) {
			found = append(found, pm)
		}
	}
	return found
}

func isInPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
