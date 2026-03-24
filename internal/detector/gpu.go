package detector

import (
	"runtime"
	"strings"
)

func DetectGPU() GPUInfo {
	switch runtime.GOOS {
	case "linux":
		return detectGPULinux()
	case "darwin":
		return detectGPUDarwin()
	default:
		return GPUInfo{Detected: false}
	}
}

func detectGPULinux() GPUInfo {
	out := runCmd("nvidia-smi", "--query-gpu=name,driver_version", "--format=csv,noheader,nounits")
	if out != "" {
		model, driver := parseNvidiaSMI(out)
		return GPUInfo{Detected: true, Model: model, Driver: driver}
	}
	out = runCmd("lspci")
	if out != "" {
		for _, line := range strings.Split(out, "\n") {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "vga") || strings.Contains(lower, "3d") || strings.Contains(lower, "display") {
				if strings.Contains(lower, "nvidia") {
					return GPUInfo{Detected: true, Model: strings.TrimSpace(line)}
				}
			}
		}
	}
	return GPUInfo{Detected: false}
}

func detectGPUDarwin() GPUInfo {
	out := runCmd("system_profiler", "SPDisplaysDataType")
	if out != "" && strings.Contains(out, "Chipset Model") {
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "Chipset Model") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					return GPUInfo{Detected: true, Model: strings.TrimSpace(parts[1])}
				}
			}
		}
	}
	return GPUInfo{Detected: false}
}

func parseNvidiaSMI(output string) (model, driver string) {
	output = strings.TrimSpace(output)
	if output == "" {
		return "", ""
	}
	parts := strings.SplitN(output, ",", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return output, ""
}
