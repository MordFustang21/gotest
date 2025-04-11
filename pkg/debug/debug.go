package debug

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/MordFustang21/gotest/pkg/pathutils"
	"github.com/MordFustang21/gotest/pkg/testutils"
)

func DebugTest(t testutils.Test, path, modRoot string) (*exec.Cmd, bool, error) {
	p, err := exec.LookPath("dlv")
	if err != nil {
		return nil, false, fmt.Errorf("error looking for dlv: %w", err)
	}

	// Create a temp file to set breakpoints and tell dlv to continue.
	tempFile, err := os.CreateTemp("", "go-test_*")
	if err != nil {
		return nil, false, fmt.Errorf("error creating temp file: %w", err)
	}

	// Attempt cleanup when no longer in use.
	defer func() {
		err = os.Remove(tempFile.Name())
		if err != nil {
			log.Println("error removing breakpoint file", err)
		}
	}()

	tempFile.Write([]byte("b " + fmt.Sprintf("%s:%d", pathutils.PackageFromPathAndMod(t.FilePath, modRoot), t.LineNumber) + "\n"))
	tempFile.Write([]byte("c\n"))
	tempFile.Close()

	args := []string{"dlv", "test", "--init", tempFile.Name(), filepath.Dir(t.FilePath), "--", "-test.run", t.Name}
	fmt.Println("Running test with debugger:", args)

	cmd := exec.Cmd{
		Path:   p,
		Env:    os.Environ(),
		Args:   args,
		Dir:    modRoot,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	err = cmd.Run()
	if err != nil {
		return nil, false, fmt.Errorf("error running dlv: %w", err)
	}

	return &cmd, true, nil
}
