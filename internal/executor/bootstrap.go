package executor

import (
	"context"
	"os/exec"
	"time"

	"github.com/host452b/isetup/internal/detector"
	"github.com/host452b/isetup/internal/logger"
)

// bootstrapPkgs are essential tools that many install scripts depend on.
// Without these, shell-based installs (curl ... | bash) will fail.
var bootstrapPkgs = []string{"curl", "wget", "ca-certificates", "gnupg"}

// Bootstrap ensures minimal prerequisites exist on the system.
// On Debian/Ubuntu, bare containers often lack curl, wget, and CA certs.
// This runs before any profile tools are installed.
func Bootstrap(ctx context.Context, info *detector.SystemInfo, lg *logger.Logger) {
	if info.OS != "linux" {
		return
	}

	// Only bootstrap if apt or apt-get is available
	aptCmd := resolveAptCmd(info)
	if aptCmd == "" {
		return
	}

	// Find which bootstrap packages are missing
	var missing []string
	for _, pkg := range bootstrapPkgs {
		if !isBootstrapInstalled(pkg) {
			missing = append(missing, pkg)
		}
	}
	if len(missing) == 0 {
		return
	}

	sudo := ""
	if !info.IsRoot {
		sudo = "sudo "
	}

	// Run apt update first
	updateCtx, updateCancel := context.WithTimeout(ctx, 2*time.Minute)
	updateResult := Run(updateCtx, sudo+aptCmd+" update", info.Shell)
	updateCancel()

	if lg != nil {
		_ = lg.WriteToolResult(logger.ToolResult{
			Name:     "_bootstrap_update",
			Profile:  "_bootstrap",
			Method:   "apt",
			Command:  sudo + aptCmd + " update",
			Status:   statusFromExit(updateResult.ExitCode),
			ExitCode: updateResult.ExitCode,
			Duration: updateResult.Duration,
			Stdout:   updateResult.Stdout,
			Stderr:   updateResult.Stderr,
		})
	}

	if updateResult.ExitCode != 0 {
		// apt update failed — try apt-get if we used apt
		if aptCmd == "apt" {
			aptCmd = "apt-get"
			retryCtx, retryCancel := context.WithTimeout(ctx, 2*time.Minute)
			retryResult := Run(retryCtx, sudo+"apt-get update", info.Shell)
			retryCancel()
			if retryResult.ExitCode != 0 {
				return // both failed, skip bootstrap
			}
		} else {
			return
		}
	}

	// Install missing packages
	for _, pkg := range missing {
		cmd := sudo + aptCmd + " install -y " + pkg
		installCtx, installCancel := context.WithTimeout(ctx, 2*time.Minute)
		result := Run(installCtx, cmd, info.Shell)
		installCancel()

		// apt → apt-get fallback
		if result.ExitCode != 0 && aptCmd == "apt" {
			fallback := sudo + "apt-get install -y " + pkg
			retryCtx, retryCancel := context.WithTimeout(ctx, 2*time.Minute)
			result = Run(retryCtx, fallback, info.Shell)
			retryCancel()
			if result.ExitCode == 0 {
				cmd = fallback
			}
		}

		if lg != nil {
			_ = lg.WriteToolResult(logger.ToolResult{
				Name:     "_bootstrap_" + pkg,
				Profile:  "_bootstrap",
				Method:   "apt",
				Command:  cmd,
				Status:   statusFromExit(result.ExitCode),
				ExitCode: result.ExitCode,
				Duration: result.Duration,
				Stdout:   result.Stdout,
				Stderr:   result.Stderr,
			})
		}
	}

	// Refresh detected package managers after bootstrap
	info.PkgManagers = detector.DetectPkgManagers()
}

// resolveAptCmd returns "apt" or "apt-get", whichever is available. Prefers apt.
func resolveAptCmd(info *detector.SystemInfo) string {
	if hasPkgMgr(info.PkgManagers, "apt") {
		return "apt"
	}
	if hasPkgMgr(info.PkgManagers, "apt-get") {
		return "apt-get"
	}
	return ""
}

// isBootstrapInstalled checks if a tool binary is in PATH.
// For packages like ca-certificates that don't have a binary, check known file paths.
func isBootstrapInstalled(pkg string) bool {
	switch pkg {
	case "ca-certificates":
		// ca-certificates doesn't provide a binary; check the cert bundle
		for _, path := range []string{
			"/etc/ssl/certs/ca-certificates.crt",
			"/etc/pki/tls/certs/ca-bundle.crt",
		} {
			if _, err := exec.LookPath(path); err == nil {
				return true
			}
		}
		// Also check if the package is installed via dpkg
		cmd := exec.Command("dpkg", "-s", "ca-certificates")
		return cmd.Run() == nil
	case "gnupg":
		_, err := exec.LookPath("gpg")
		return err == nil
	default:
		_, err := exec.LookPath(pkg)
		return err == nil
	}
}

func statusFromExit(code int) string {
	if code == 0 {
		return logger.StatusSuccess
	}
	return logger.StatusFailed
}
