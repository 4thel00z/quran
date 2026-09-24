package font

import (
	"os/exec"
	"runtime"
)

// RefreshCache invokes the system font cache updater if available (e.g. fc-cache on Linux/BSD).
// Returns a status message and error if execution failed.
func RefreshCache(targetDir string) (string, error) {
	if runtime.GOOS != "linux" && runtime.GOOS != "freebsd" && runtime.GOOS != "openbsd" && runtime.GOOS != "netbsd" {
		return "Font cache refresh is not required on this OS.", nil
	}

	cmdPath, err := exec.LookPath("fc-cache")
	if err != nil {
		return "fc-cache not found in PATH; font cache not refreshed.", nil
	}

	cmd := exec.Command(cmdPath, "-f", targetDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return string(out), err
	}

	return "Font cache successfully refreshed with fc-cache.", nil
}
