package benchmark

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/MordFustang21/gotest/pkg/flamegraph"
	"github.com/MordFustang21/gotest/pkg/history"
	"github.com/MordFustang21/gotest/pkg/pathutils"
	"github.com/MordFustang21/gotest/pkg/testutils"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

func RunBenchmark(t testutils.Test) error {
	path, modRoot, err := pathutils.TestToPathAndModRoot(t)
	if err != nil {
		return fmt.Errorf("error getting path to test: %w", err)
	}

	// create base args with verbose and a run that filters tests so we only run benchmarks
	args := []string{"test", "-v", path, "-run", "XXX"}
	if t.Name != "" {
		args = append(args, "-bench", t.Name)
	}

	var cpuProfile string
	if viper.GetBool("cpu") {
		tempFile, err := os.CreateTemp("", "go-test_"+t.Name)
		if err != nil {
			return fmt.Errorf("error creating CPU profile file: %w", err)
		}

		cpuProfile = tempFile.Name()
		tempFile.Close()

		args = append(args, "-cpuprofile", cpuProfile)
	}

	var memoryProfile string
	if viper.GetBool("mem") {
		tempFile, err := os.CreateTemp("", "go-test_"+t.Name)
		if err != nil {
			return fmt.Errorf("error creating memory profile file: %w", err)
		}

		memoryProfile = tempFile.Name()
		tempFile.Close()

		args = append(args, "-memprofile", memoryProfile)
	}

	p, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("error finding go binary: %w", err)
	}

	// create a buffer to capture the output of the benchmark
	benchBuffer := &bytes.Buffer{}
	writer := io.MultiWriter(os.Stdout, benchBuffer)

	cmd := exec.Cmd{
		Path:   p,
		Env:    os.Environ(),
		Args:   append([]string{"go"}, args...),
		Dir:    modRoot,
		Stdout: writer,
		Stderr: os.Stderr,
	}

	fmt.Println("Running", cmd.Args, "@", cmd.Dir)

	err = cmd.Run()
	switch {
	case err == nil:
		// store the successful benchmark
		storeBenchmarkResult(cmd, benchBuffer)
	case errors.Is(err, &exec.ExitError{}):
	// do nothing
	default:
		return fmt.Errorf("error running benchmark: %w", err)
	}

	if viper.GetBool("cpu") {
		fmt.Println("Wrote CPU Profile to:", cpuProfile)
		// Generate the flamegraph
		svgData, err := flamegraph.GenerateFlamegraph(cpuProfile)
		if err != nil {
			return fmt.Errorf("error generating flamegraph: %w", err)
		}

		err = flamegraph.ServeFlamegraph(svgData)
		if err != nil {
			panic(err)
		}
	}

	if viper.GetBool("mem") {
		fmt.Println("Wrote Memory Profile to:", memoryProfile)
		cmd := exec.Command("go", "tool", "pprof", "-top", memoryProfile)
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("error running memory profile: %w", err)
		}
	}

	return nil
}

const benchmarkDB = "benchmarks.db"

func storeBenchmarkResult(cmd exec.Cmd, benchBuffer *bytes.Buffer) {
	db := history.GetHistoryFile(benchmarkDB)

	db.Update(func(tx *bolt.Tx) error {
		return nil
	})
}
