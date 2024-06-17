package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed wasm_exec_node.js
var wasmExecNode []byte

//go:embed wasm_exec.js
var wasmExec []byte

func main() {
	// find wasm dir
	dir := filepath.Join(runtime.GOROOT(), "misc", "wasm")

	if d, err := os.Stat(dir); err != nil || !d.IsDir() {
		// use embedded one
		// create a temp dir
		dir, err = os.UserCacheDir()
		if err != nil {
			panic(err)
		}
		dir = filepath.Join(dir, "wasm_exec")
		if d, err := os.Stat(dir); os.IsNotExist(err) || !d.IsDir() {
			if err := os.MkdirAll(dir, 0755); err != nil {
				panic(err)
			}

			if err := os.WriteFile(filepath.Join(dir, "wasm_exec_node.js"), wasmExecNode, 0644); err != nil {
				panic(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "wasm_exec.js"), wasmExec, 0644); err != nil {
				panic(err)
			}
		}

		_, _ = fmt.Fprintln(os.Stderr, "Warning: Can not read from go installation, use embedded file at "+dir)
	}

	args := []string{"--stack-size=8192", "wasm_exec_node.js"}
	args = append(args, os.Args[1:]...)

	node, err := exec.LookPath("node")
	if err != nil {
		panic(fmt.Errorf("node not found in PATH: %w", err))
	}

	cmd := exec.Command(node, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	//calculate env size
	envSize := 0
	for _, e := range os.Environ() {
		envSize += len(e) + 1
	}
	if envSize > 8192 {
		_, _ = fmt.Fprintln(os.Stderr, "Warning: env size is too large, auto filter the PATH env")
		newEnv := make([]string, 0)
		for _, e := range os.Environ() {
			name := strings.Split(e, "=")[0]
			if strings.EqualFold(name, "PATH") {
				continue
			}
		}
		cmd.Env = newEnv
	}

	// filter env
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
