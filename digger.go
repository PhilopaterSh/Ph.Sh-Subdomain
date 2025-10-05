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
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	scriptPath := filepath.Join(exeDir, "digger_wrapper.py")
	// fmt.Printf("DEBUG: scriptPath = %s\n", scriptPath)


	// Check if the script exists, if not, create it from the embedded content.
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		// fmt.Printf("DEBUG: digger_wrapper.py does not exist at %s, creating it.\n", scriptPath)
		err := os.WriteFile(scriptPath, []byte(diggerWrapperScript), 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create digger_wrapper.py: %v", err)
		}
	} else {
		// fmt.Printf("DEBUG: digger_wrapper.py already exists at %s.\n", scriptPath)
	}

	cmd := exec.Command("python", scriptPath, domain)

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