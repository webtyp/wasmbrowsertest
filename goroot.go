package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// goRoot resolves GOROOT by asking the installed toolchain. runtime.GOROOT is the root the binary was built with,
// which means nothing once the binary is copied to another machine.
func goRoot() (string, error) {
	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return "", fmt.Errorf("go env GOROOT failed: %w", err)
	}

	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", fmt.Errorf("go env GOROOT returned an empty path")
	}
	return root, nil
}
