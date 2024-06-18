package main

import (
	"bytes"
	"errors"
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

	isToolchain := IsToolChain()

	if isToolchain {
		version = ReadVersion()
		cacheDir := Must2(os.UserCacheDir())

		dir = filepath.Join(cacheDir, "wasm-exec", version)
		Must(os.MkdirAll(dir, 0755))
	} else {
		dir = filepath.Join(runtime.GOROOT(), "misc", "wasm")
	}
	Must(RequireFile(dir, "wasm_exec.js", version, isToolchain))
	Must(RequireFile(dir, "wasm_exec_node.js", version, isToolchain))

	return dir
}

func IsToolChain() bool {
	goroot := runtime.GOROOT()
	return strings.Contains(goroot, "golang.org/toolchain") ||
		(strings.Contains(goroot, "golang.org\\toolchain"))
}

func ReadVersion() string {
	versionFile := Must2(os.ReadFile(filepath.Join(runtime.GOROOT(), "VERSION")))
	lines := strings.Split(string(versionFile), "\n")
	return lines[0]
}

func RequireFile(dir string, filename string, version string, tryDownload bool) error {
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

	if !tryDownload {
		panic(errors.New("file not found: " + absFileName))
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/golang/go/%s/misc/wasm/%s", version, filename)

	resp := Must2(http.Get(url))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get %s failed: %s", url, resp.Status)
	}

	b := new(bytes.Buffer)

	_ = Must2(io.Copy(b, resp.Body))

	Must(os.WriteFile(absFileName, b.Bytes(), 0644))

	return nil
}
