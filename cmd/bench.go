package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/MordFustang21/gotest/pkg/benchmark"
	"github.com/MordFustang21/gotest/pkg/testutils"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(benchCmd)
}

var benchCmd = &cobra.Command{
	Use:   "bench [directory]",
	Short: "Run Go benchmarks interactively or all in a directory",
	Long: `Runs Go benchmarks found in the specified directory (or current directory).
Allows interactive selection of a specific benchmark function to run.
Supports profiling flags (--cpu, --mem).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		readDir, _ := os.Getwd()
		if len(args) > 0 {
			readDir = args[0] // Use provided directory argument
		}

		// Find available benchmarks
		availableBenchmarks, err := testutils.GetTestsFromDir(readDir, true) // true for benchmarksOnly
		if err != nil {
			return fmt.Errorf("error getting benchmarks from %s: %w", readDir, err)
		}

		if len(availableBenchmarks) == 0 {
			fmt.Println("No benchmarks found in the directory:", readDir)
			return nil
		}

		// Select benchmark interactively
		selectedBench, err := testutils.SelectTest(availableBenchmarks)
		if err != nil {
			if errors.Is(err, promptui.ErrInterrupt) {
				fmt.Println("Benchmark selection cancelled.")
				return nil
			}
			return fmt.Errorf("error selecting benchmark: %w", err)
		}

		fmt.Printf("Running benchmark: %s in %s\n", selectedBench.Name, selectedBench.File)
		err = benchmark.RunBenchmark(selectedBench)
		if err != nil {
			// benchmark.RunBenchmark should ideally return a specific error
			// type if the benchmark itself failed, vs. setup errors.
			return fmt.Errorf("benchmark execution failed: %w", err)
		}

		return nil
	},
}
