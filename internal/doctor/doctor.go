// Package doctor checks for required tools and suggests installation commands.
package doctor

import (
	"os/exec"
	"runtime"
	"strings"
)

type Tool struct {
	Cmd  string
	Args []string
}

// Tools lists the tools checked by `devkit doctor`.
var Tools = []Tool{
	{"git", []string{"--version"}},
	{"docker", []string{"--version"}},
	{"python3", []string{"--version"}},
	{"node", []string{"--version"}},
	{"go", []string{"version"}},
	{"java", []string{"-version"}},
}

var hints = map[string]map[string]string{
	"git":     {"darwin": "brew install git", "linux": "sudo apt install git", "windows": "winget install Git.Git"},
	"docker":  {"darwin": "brew install --cask docker", "linux": "https://docs.docker.com/engine/install/", "windows": "winget install Docker.DockerDesktop"},
	"python3": {"darwin": "brew install python", "linux": "sudo apt install python3 python3-venv", "windows": "winget install Python.Python.3.12"},
	"node":    {"darwin": "brew install node", "linux": "sudo apt install nodejs npm", "windows": "winget install OpenJS.NodeJS.LTS"},
	"go":      {"darwin": "brew install go", "linux": "sudo apt install golang-go", "windows": "winget install GoLang.Go"},
	"java":    {"darwin": "brew install openjdk", "linux": "sudo apt install default-jdk", "windows": "winget install EclipseAdoptium.Temurin.21.JDK"},
}

func binName(cmd string) string {
	if runtime.GOOS == "windows" && cmd == "python3" {
		return "python"
	}
	return cmd
}

// Version returns the first line of the version command's output.
func Version(t Tool) (string, bool) {
	path, err := exec.LookPath(binName(t.Cmd))
	if err != nil {
		return "", false
	}
	out, _ := exec.Command(path, t.Args...).CombinedOutput()
	line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return line, true
}

func Has(cmd string) bool {
	_, err := exec.LookPath(binName(cmd))
	return err == nil
}

func Hint(cmd string) string { return hints[cmd][runtime.GOOS] }
