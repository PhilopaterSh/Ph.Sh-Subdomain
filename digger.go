package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed digger_wrapper.py
var diggerWrapperScript string

// DiggerToolsEngine is an engine that uses a Python script to query Digger.tools.
type DiggerToolsEngine struct{}

// Name returns the name of the engine.
func (e *DiggerToolsEngine) Name() string {
	return "digger"
}

// Fetch returns a list of subdomains found by the engine.
func (e *DiggerToolsEngine) Fetch(domain string, client *http.Client) ([]string, error) {
	// Write the embedded script to a user-writable cache directory (not next to
	// the binary, which may live in a protected path like /usr/local/bin).
	baseDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user cache directory: %v", err)
	}
	scriptDir := filepath.Join(baseDir, "Ph.Sh_Sub")
	scriptPath := filepath.Join(scriptDir, "digger_wrapper.py")

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		if err := os.MkdirAll(scriptDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create digger cache directory: %v", err)
		}
	}
	// Keep the script in sync with the binary: rewrite it whenever the embedded
	// content changed (e.g. after a tool upgrade).
	if existing, err := os.ReadFile(scriptPath); err != nil || string(existing) != diggerWrapperScript {
		if err := os.WriteFile(scriptPath, []byte(diggerWrapperScript), 0755); err != nil {
			return nil, fmt.Errorf("failed to create digger_wrapper.py: %v", err)
		}
	}

	// Prefer python3 (the only interpreter present on modern Linux/macOS).
	python, err := exec.LookPath("python3")
	if err != nil {
		python, err = exec.LookPath("python")
		if err != nil {
			return nil, fmt.Errorf("no python interpreter found in PATH (tried python3 and python)")
		}
	}

	cmd := exec.Command(python, scriptPath, domain)

	output, err := cmd.Output()
	if err != nil {
		// If the python script itself fails, return an error with details
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("digger_wrapper.py failed with exit code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
		}
		return nil, fmt.Errorf("failed to run digger_wrapper.py: %v", err)
	}

	subs := strings.Split(strings.TrimSpace(string(output)), "\n")
	// If the output is empty, return an empty slice
	if len(subs) == 1 && subs[0] == "" {
		return []string{}, nil
	}

	return subs, nil
}
