package install

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// DefaultBinDir prefers ~/bin, else GOPATH/bin.
func DefaultBinDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	bin := filepath.Join(home, "bin")
	if st, err := os.Stat(bin); err == nil && st.IsDir() {
		return bin, nil
	}
	if gopath := strings.TrimSpace(os.Getenv("GOPATH")); gopath != "" {
		return filepath.Join(strings.Split(gopath, string(os.PathListSeparator))[0], "bin"), nil
	}
	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err == nil {
		gp := strings.TrimSpace(string(out))
		if gp != "" {
			return filepath.Join(gp, "bin"), nil
		}
	}
	return bin, nil
}

// CopyExecutable copies the current process binary to destPath (dest is full path including "bay").
func CopyExecutable(destPath string) error {
	src, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	src, err = filepath.EvalSymlinks(src)
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := destPath + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, destPath); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// PathContainsDir reports whether dir appears in PATH.
func PathContainsDir(dir string) bool {
	dir = filepath.Clean(dir)
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if filepath.Clean(p) == dir {
			return true
		}
	}
	return false
}

// PathHint returns a shell snippet to prepend dir to PATH (zsh/bash).
func PathHint(dir string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`Add "%s" to your PATH environment variable.`, dir)
	}
	return fmt.Sprintf(`export PATH="%s:$PATH"`, dir)
}
