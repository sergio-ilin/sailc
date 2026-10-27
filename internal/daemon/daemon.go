package daemon

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/windows/registry"
)

const (
	RegistryRunKey = `Software\Microsoft\Windows\CurrentVersion\Run`
	AppName        = "SailC_Daemon"
)

// IsRunning checks whether the background MCP hub is active on the given port.
func IsRunning(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 250*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		return true
	}
	return false
}

// GetDaemonPath resolves the absolute path to sailcd.exe.
func GetDaemonPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return "sailcd.exe"
	}
	dir := filepath.Dir(exePath)
	daemonPath := filepath.Join(dir, "sailcd.exe")
	if _, err := os.Stat(daemonPath); err == nil {
		return daemonPath
	}
	return exePath
}

// EnsureRunning guarantees that the background MCP daemon process is running.
// If not active, it launches sailcd.exe completely detached
// without creating any console window (CREATE_NO_WINDOW | DETACHED_PROCESS).
func EnsureRunning(port int) error {
	if IsRunning(port) {
		return nil
	}

	daemonPath := GetDaemonPath()
	cmd := exec.Command(daemonPath, fmt.Sprintf("-port=%d", port))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000 | 0x00000008, // CREATE_NO_WINDOW | DETACHED_PROCESS
		HideWindow:    true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Poll until ready (up to 3 seconds)
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if IsRunning(port) {
			return nil
		}
	}

	return nil
}

// InstallStartup registers the background daemon in Windows HKCU Run key for 24/7 autostart on user login.
func InstallStartup(port int) error {
	daemonPath := GetDaemonPath()

	k, err := registry.OpenKey(registry.CURRENT_USER, RegistryRunKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	cmdLine := fmt.Sprintf(`"%s" -port=%d`, daemonPath, port)
	return k.SetStringValue(AppName, cmdLine)
}

// UninstallStartup removes the daemon from Windows HKCU Run key.
func UninstallStartup() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, RegistryRunKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	return k.DeleteValue(AppName)
}

// AttachParentConsole attaches to the parent console if launched from a terminal.
func AttachParentConsole() {
	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole := modkernel32.NewProc("AttachConsole")
	_, _, _ = procAttachConsole.Call(uintptr(^uint32(0))) // ATTACH_PARENT_PROCESS = (DWORD)-1
}
