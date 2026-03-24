package detector

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func ArchLabel(goarch, goos string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	case "arm64":
		if goos == "darwin" {
			return "arm64"
		}
		return "aarch64"
	default:
		return goarch
	}
}

func DetectOS() *SystemInfo {
	info := &SystemInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		ArchLabel: ArchLabel(runtime.GOARCH, runtime.GOOS),
		Home:      os.Getenv("HOME"),
	}

	switch runtime.GOOS {
	case "linux":
		info.Distro = detectLinuxDistro()
		info.Kernel = runCmd("uname", "-r")
		info.WSL = detectWSL()
	case "darwin":
		info.Distro = detectMacOSVersion()
		info.Kernel = runCmd("uname", "-r")
	case "windows":
		info.Home = os.Getenv("USERPROFILE")
		info.Distro = runCmd("cmd", "/c", "ver")
		info.Kernel = info.Distro
	}

	return info
}

func detectLinuxDistro() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			val := strings.TrimPrefix(line, "PRETTY_NAME=")
			return strings.Trim(val, "\"")
		}
	}
	return ""
}

func detectMacOSVersion() string {
	name := runCmd("sw_vers", "-productName")
	ver := runCmd("sw_vers", "-productVersion")
	if name != "" && ver != "" {
		return name + " " + ver
	}
	return ""
}

func detectWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

func runCmd(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
