package detector

type SystemInfo struct {
	OS                string   `json:"os"`
	Arch              string   `json:"arch"`
	ArchLabel         string   `json:"arch_label"`
	Distro            string   `json:"distro"`
	Kernel            string   `json:"kernel"`
	WSL               bool     `json:"wsl"`
	Shell             string   `json:"shell"`
	PowerShellVersion string   `json:"powershell_version"`
	GPU               GPUInfo  `json:"gpu"`
	PkgManagers       []string `json:"pkg_managers"`
	Home              string   `json:"home"`
}

type GPUInfo struct {
	Detected bool   `json:"detected"`
	Model    string `json:"model,omitempty"`
	Driver   string `json:"driver,omitempty"`
}

func Detect() *SystemInfo {
	info := DetectOS()
	info.GPU = DetectGPU()
	info.PkgManagers = DetectPkgManagers()
	shell, psVer := DetectShell()
	info.Shell = shell
	info.PowerShellVersion = psVer
	return info
}
