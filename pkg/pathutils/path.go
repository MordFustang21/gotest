package pathutils

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/MordFustang21/gotest/pkg/testutils"
)

// TestToPathAndModRoot converts a test case to a relative path and finds the module root.
func TestToPathAndModRoot(t testutils.Test) (path, modRoot string, err error) {
	path = t.FilePath

	// if this is a file and not a directory, use the directory of the file
	if filepath.Ext(t.FilePath) != "" {
		path = filepath.Dir(t.FilePath)
	}

	modRoot = LookupModuleRoot(path)
	if modRoot == "" {
		return "", "", errors.New("could not find module root for path: " + path)
	}

	// convert path to a module path
	path, _ = filepath.Rel(modRoot, path)
	path = "./" + path

	// in the event the directory is the root of the module, and there isn't a named test specified we need to add an extra ".." to
	// tell go test to recursively run all tests
	if path == "./." {
		path += ".."
	}

	return path, modRoot, nil
}

func LookupModuleRoot(path string) string {
	// start at end and work backwards to find the go.mod file
	for {
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			// Return the full root path if go.mod is found
			absPath, err := filepath.Abs(path)
			if err != nil {
				panic("could not get absolute path for module root: " + err.Error())
			}

			return absPath
		}

		path = filepath.Dir(path)

		if path == "/" {
			break
		}
	}

	return ""
}

func ResolvePackage(path, modRoot string) string {
	cmd := exec.Command("go", "list", "-f", "{{.ImportPath}}", path)
	cmd.Dir = modRoot
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	return string(bytes.TrimSpace(out))
}

func PackageFromPathAndMod(path, modRoot string) string {
	out, err := filepath.Rel(modRoot, path)
	if err != nil {
		return ""
	}

	return out
}
