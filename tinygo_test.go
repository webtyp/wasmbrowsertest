package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"webtyp.com/tinygo"
)

func TestTinygoWasmExecJS(t *testing.T) {
	if !tinygo.IsInstalled() {
		t.Skip("tinygo is not installed")
	}

	buf, err := tinygoWasmExecJS()
	if err != nil {
		t.Fatalf("tinygoWasmExecJS failed: %v", err)
	}
	if len(buf) == 0 {
		t.Fatal("tinygoWasmExecJS returned an empty shim")
	}

	// The whole point of -tinygo: TinyGo's glue is a different file from the Go
	// toolchain's. If these ever match, serving either one would work and the
	// flag would be pointless — so a match means the resolution is wrong.
	for _, loc := range wasmLocations {
		goShim, err := os.ReadFile(filepath.Join(runtime.GOROOT(), loc))
		if err != nil {
			continue
		}
		if bytes.Equal(buf, goShim) {
			t.Fatalf("TinyGo shim is identical to the Go toolchain's at %s: wrong file resolved", loc)
		}
	}
}

func TestTinygoRootNotEmpty(t *testing.T) {
	if !tinygo.IsInstalled() {
		t.Skip("tinygo is not installed")
	}

	root, err := tinygoRoot()
	if err != nil {
		t.Fatalf("tinygoRoot failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(tinygoWasmExecLocation))); err != nil {
		t.Fatalf("shim missing under TINYGOROOT %q: %v", root, err)
	}
}
