package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MordFustang21/gotest/pkg/colorize"
	"github.com/MordFustang21/gotest/pkg/debug"
	"github.com/MordFustang21/gotest/pkg/flamegraph"
	"github.com/MordFustang21/gotest/pkg/history"
	"github.com/MordFustang21/gotest/pkg/pathutils"
	"github.com/MordFustang21/gotest/pkg/testutils"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test [directory]",
	Short: "Run Go tests in a specified directory or select interactively",
	Long: `Runs Go tests. By default, runs all tests in the target directory.
Use the --subtest (-s) flag to interactively select a specific test function.
Supports coverage, profiling, and debugging options.`,
	Args: cobra.MaximumNArgs(1), // Allow optional directory argument
	RunE: runTestCommand,
}

// runTestCommand contains the core logic for running tests
func runTestCommand(cmd *cobra.Command, args []string) error {
	if isRerun, _ := cmd.Flags().GetBool("rerun"); isRerun {
		return rerunLastTest() // Handle re-running last test command
	}

	readDir, _ := os.Getwd()
	if len(args) > 0 {
		readDir = args[0] // Use provided directory argument
	}

	isSubtest, _ := cmd.Flags().GetBool("subtest")
	isNamedTest, _ := cmd.Flags().GetString("name")

	var testToRun testutils.Test
	switch {
	case isSubtest:
		availableTests, err := testutils.GetTestsFromDir(readDir, false)
		if err != nil {
			return fmt.Errorf("error getting tests: %w", err)
		}

		if len(availableTests) == 0 {
			fmt.Println("No tests found in the directory:", readDir)
			return nil
		}

		selectedTest, err := testutils.SelectTest(availableTests)
		switch {
		case err == nil:
		case errors.Is(err, promptui.ErrInterrupt):
			fmt.Println("Test selection cancelled.")
			return nil
		default:
			return fmt.Errorf("error selecting test: %w", err)
		}

		testToRun = selectedTest
	case isNamedTest != "":
		availableTests, err := testutils.GetTestsFromDir(readDir, false)
		if err != nil {
			return fmt.Errorf("error getting tests: %w", err)
		}

		if len(availableTests) == 0 {
			fmt.Println("No tests found in the directory:", readDir)
			return nil
		}

		// Find the test with the specified name
		for _, t := range availableTests {
			if t.Name == isNamedTest {
				testToRun = t
				break
			}
		}
	default:
		// Run all tests in the directory
		absPath, err := filepath.Abs(readDir)
		if err != nil {
			return fmt.Errorf("error getting absolute path for %s: %w", readDir, err)
		}

		testToRun = testutils.Test{FilePath: absPath} // Represents running all tests in dir
	}

	execCmd, pass, err := executeTests(testToRun, cmd)
	if err != nil {
		// Don't log history if execution setup failed
		return fmt.Errorf("error executing test: %w", err)
	}

	// Log regardless of pass/fail, history needs the command attempted
	errLog := history.LogRunHistory(*execCmd, pass)
	if errLog != nil {
		// Log the error but don't fail the main command run because of history logging issue
		fmt.Fprintln(os.Stderr, "Warning: failed to log run history:", errLog)
	}

	return nil
}

func rerunLastTest() error {
	lastEntry, err := history.GetLastCommand()
	if err != nil {
		// Handle case where history is empty or file error
		return fmt.Errorf("could not get last command: %w", err)
	}

	if lastEntry == nil {
		fmt.Println("No command history found to re-run.")
		return nil
	}

	fmt.Println("Re-running last command from history:")
	// Again, RunHistoryEntry handles the execution logic
	err = history.RunHistoryEntry(*lastEntry)
	if err != nil {
		return fmt.Errorf("failed to re-run last command: %w", err)
	}

	return nil
}

// executeTests builds and runs the `go test` command based on flags
func executeTests(t testutils.Test, cmd *cobra.Command) (*exec.Cmd, bool, error) {
	path, modRoot, err := pathutils.TestToPathAndModRoot(t)
	if err != nil {
		return nil, false, fmt.Errorf("failed to determine test path: %w", err)
	}

	args := []string{"test", path}

	// Get flag values (using Viper is cleaner if bound)
	verboseFlag, _ := cmd.Flags().GetBool("verbose") // Or use rootCmd.PersistentFlags() if defined there
	debugFlag := viper.GetBool("debug")              // Assumes bound via BindPFlag
	coverFlag := viper.GetBool("cover")
	cpuFlag := viper.GetBool("cpu")
	memFlag := viper.GetBool("mem")

	if verboseFlag {
		args = append(args, "-v")
	}

	if t.Name != "" { // t.Name is set if a specific subtest was selected
		args = append(args, "-run", t.Name) // Anchor regex for exact match
	}

	if debugFlag {
		return debug.DebugTest(t, path, modRoot)
	}

	var coverFile, cpuProfile, memoryProfile string
	var cleanupFuncs []func()
	defer func() {
		for _, f := range cleanupFuncs {
			f()
		}
	}()

	if coverFlag {
		tempFile, err := os.CreateTemp("", "gotest_cover_*.out")
		if err != nil {
			return nil, false, fmt.Errorf("error creating coverage file: %w", err)
		}

		coverFile = tempFile.Name()
		tempFile.Close() // Close file handle immediately
		args = append(args, "-coverprofile="+coverFile)
		cleanupFuncs = append(cleanupFuncs, func() { os.Remove(coverFile) }) // Schedule cleanup
	}

	if cpuFlag {
		tempFile, err := os.CreateTemp("", "gotest_cpu_*.prof")
		if err != nil {
			return nil, false, fmt.Errorf("error creating CPU profile file: %w", err)
		}

		cpuProfile = tempFile.Name()
		tempFile.Close()
		args = append(args, "-cpuprofile="+cpuProfile)
		// No cleanup for profile files by default, user might want them.
	}

	if memFlag {
		tempFile, err := os.CreateTemp("", "gotest_mem_*.prof")
		if err != nil {
			return nil, false, fmt.Errorf("error creating memory profile file: %w", err)
		}

		memoryProfile = tempFile.Name()
		tempFile.Close()
		args = append(args, "-memprofile="+memoryProfile)
		// No cleanup for profile files.
	}

	// Use viper for colorization config
	colorEnabled := viper.GetBool("colorizeoutput")

	var outputWriter io.Writer = os.Stdout
	if colorEnabled {
		var colorReader io.Reader
		colorReader, outputWriter = io.Pipe()
		go colorize.ColorizeOutput(colorReader, os.Stdout)
	}

	// Look up the 'go' executable in PATH, exec.Cmd doesn't automatically resolve it.
	goPath, err := exec.LookPath("go")
	if err != nil {
		return nil, false, fmt.Errorf("error looking up go executable: %w", err)
	}

	execCmd := exec.Cmd{
		Path:   goPath,
		Args:   append([]string{"go"}, args...),
		Env:    os.Environ(),
		Dir:    modRoot,
		Stdout: outputWriter,
		Stderr: os.Stderr,
	}

	fmt.Println("Running:", strings.Join(execCmd.Args, " "), "@", execCmd.Dir)

	var pass bool
	err = execCmd.Run()
	var exit *exec.ExitError
	switch {
	case err == nil:
		pass = true
	case errors.As(err, &exit):
		// do nothing
	default:
		return nil, false, fmt.Errorf("error running command: %w", err)
	}

	if coverFlag && coverFile != "" {
		fmt.Println("Coverage profile:", coverFile)
		coverCmd := exec.Command(goPath, "tool", "cover", "-html="+coverFile)
		coverCmd.Stdout = os.Stdout
		coverCmd.Stderr = os.Stderr
		coverCmd.Dir = modRoot // Run from module root
		if err := coverCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Warning: failed to launch coverage viewer:", err)
			// Don't fail the whole command for this
		}
	}

	if cpuFlag && cpuProfile != "" {
		fmt.Println("CPU Profile:", cpuProfile)
		svgData, err := flamegraph.GenerateFlamegraph(cpuProfile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: error generating CPU flamegraph:", err)
		} else {
			err = flamegraph.ServeFlamegraph(svgData) // This might block
			if err != nil {
				fmt.Fprintln(os.Stderr, "Warning: error serving CPU flamegraph:", err)
			}
		}
	}

	if memFlag && memoryProfile != "" {
		fmt.Println("Memory Profile:", memoryProfile)
		pprofCmd := exec.Command(goPath, "tool", "pprof", "-top", memoryProfile)
		pprofCmd.Stdout = os.Stdout
		if err := pprofCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Warning: error running memory profile analysis:", err)
		}
	}

	// Note: runErr is handled above, we return nil error from this function if execution was possible
	return &execCmd, pass, nil
}
