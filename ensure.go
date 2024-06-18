package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must2[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func RequireValidWasmDir() string {
	var dir string
	var version string
	if IsToolChain() {
		version = ReadVersion()
		cacheDir := Must2(os.UserCacheDir())

		dir = filepath.Join(cacheDir, "wasm-exec", version)
		Must(os.MkdirAll(dir, 0755))
	} else {
		dir = runtime.GOROOT()
	}
	Must(RequireFile(dir, "misc/wasm/wasm_exec.js", version))
	Must(RequireFile(dir, "misc/wasm/wasm_exec_node.js", version))

	return filepath.Join(dir, "misc/wasm")
}

func IsToolChain() bool {
	return strings.Contains(runtime.GOROOT(), "golang.org"+string(os.PathSeparator)+"toolchain")
}

func ReadVersion() string {
	versionFile := Must2(os.ReadFile(filepath.Join(runtime.GOROOT(), "VERSION")))
	lines := strings.Split(string(versionFile), "\n")
	return lines[0]
}

func RequireFile(dir string, filename string, version string) error {
	absFileName := filepath.Join(dir, filename)

	baseDir := filepath.Dir(absFileName)
	d, err := os.Stat(baseDir)
	if os.IsNotExist(err) {
		Must(os.MkdirAll(baseDir, 0755))
	} else {
		if !d.IsDir() {
			return fmt.Errorf("dir %s is not a directory", baseDir)
		}
	}

	_, err = os.Stat(absFileName)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/golang/go/%s/%s", version, filename)

	f := Must2(os.Create(absFileName))
	defer func() {
		Must(f.Close())
	}()

	resp := Must2(http.Get(url))

	_ = Must2(io.Copy(f, resp.Body))

	return nil
}
