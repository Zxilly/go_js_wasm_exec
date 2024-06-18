package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	args := []string{"--stack-size=8192", "wasm_exec_node.js"}
	args = append(args, os.Args[1:]...)

	node, err := exec.LookPath("node")
	if err != nil {
		panic(fmt.Errorf("node not found in PATH: %w", err))
	}

	cmd := exec.Command(node, args...)
	cmd.Dir = RequireValidWasmDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	//calculate env size
	envSize := 0
	for _, e := range os.Environ() {
		envSize += len(e) + 1
	}
	if envSize > 8192 {
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
