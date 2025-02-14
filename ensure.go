package main

import (
	"errors"
	"fmt"
	"go/version"
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
	var ver string

	shouldDownload := IsToolChain() && version.Compare(runtime.Version(), "go1.24") < 0

	if shouldDownload {
		ver = ReadVersion()
		cacheDir := Must2(os.UserCacheDir())

		dir = filepath.Join(cacheDir, "wasm-exec", ver)
		Must(os.MkdirAll(dir, 0755))
	} else {
		if version.Compare(runtime.Version(), "go1.24") < 0 {
			dir = filepath.Join(runtime.GOROOT(), "misc", "wasm")
		} else {
			dir = filepath.Join(runtime.GOROOT(), "lib", "wasm")
		}
	}
	Must(RequireFile(dir, "wasm_exec.js", ver, shouldDownload))
	Must(RequireFile(dir, "wasm_exec_node.js", ver, shouldDownload))

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

	const base = "https://raw.githubusercontent.com/golang/go"

	url := fmt.Sprintf(base+"/%s/misc/wasm/%s", version, filename)

	resp := Must2(http.Get(url))
	defer func(Body io.ReadCloser) {
		Must(Body.Close())
	}(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		url = fmt.Sprintf(base+"/%s/lib/wasm/%s", version, filename)
		resp = Must2(http.Get(url))
		defer func(Body io.ReadCloser) {
			Must(Body.Close())
		}(resp.Body)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get %s failed: %s", url, resp.Status)
	}

	Must(os.WriteFile(absFileName, Must2(io.ReadAll(resp.Body)), 0644))

	return nil
}
