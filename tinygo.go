package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"webtyp.com/tinygo"
)

// tinygoWasmExecLocation is the shim's path relative to TINYGOROOT.
const tinygoWasmExecLocation = "targets/wasm_exec.js"

// tinygoTarget is the TinyGo build target for the browser.
const tinygoTarget = "wasm"

// tinygoBuildTest recompiles the package under test with TinyGo and returns the
// resulting wasm binary.
//
// `go test -exec` hands us a binary the *Go* toolchain already produced, which
// tells us nothing about TinyGo compatibility: Go's js/wasm backend supports the
// full stdlib, so a package that Go compiles happily can still be rejected by
// TinyGo. To actually exercise TinyGo we discard that binary and rebuild from
// source. `tinygo test` has no -exec hook of its own, hence -c.
//
// `go test -exec` runs us with the package's source directory as the working
// directory, so "." is the package under test.
func tinygoBuildTest() (wasmFile string, cleanup func(), err error) {
	bin, err := tinygoBin()
	if err != nil {
		return "", nil, err
	}

	tmpDir, err := os.MkdirTemp("", "wasmbrowsertest-tinygo-*")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { os.RemoveAll(tmpDir) }

	out := filepath.Join(tmpDir, "pkg.wasm")
	cmd := exec.Command(bin, "test", "-target", tinygoTarget, "-c", "-o", out, ".")

	if combined, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("tinygo build failed:\n%s", combined)
	}
	return out, cleanup, nil
}

// tinygoWasmExecJS returns TinyGo's JS glue.
//
// TinyGo's wasm_exec.js is NOT interchangeable with the Go toolchain's: the two
// compilers emit different host-import signatures, so serving GOROOT's shim to a
// TinyGo binary fails at instantiation with an opaque link error. It ships inside
// TINYGOROOT, which is why the toolchain has to be located first.
func tinygoWasmExecJS() ([]byte, error) {
	root, err := tinygoRoot()
	if err != nil {
		return nil, err
	}

	shim := filepath.Join(root, filepath.FromSlash(tinygoWasmExecLocation))
	buf, err := os.ReadFile(shim)
	if err != nil {
		return nil, fmt.Errorf("tinygo: cannot read %s: %w", shim, err)
	}
	return buf, nil
}

// tinygoBin locates the TinyGo toolchain, in PATH or locally installed.
func tinygoBin() (string, error) {
	bin, err := tinygo.GetPath()
	if err != nil {
		return "", fmt.Errorf("tinygo is not installed: %w\n"+
			"install it with: go run webtyp.com/tinygo/cmd/tinygoinstall@latest", err)
	}
	return bin, nil
}

// tinygoRoot resolves TINYGOROOT by asking the installed toolchain.
func tinygoRoot() (string, error) {
	bin, err := tinygoBin()
	if err != nil {
		return "", err
	}

	out, err := exec.Command(bin, "env", "TINYGOROOT").Output()
	if err != nil {
		return "", fmt.Errorf("tinygo env TINYGOROOT failed: %w", err)
	}

	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", fmt.Errorf("tinygo env TINYGOROOT returned an empty path")
	}
	return root, nil
}
