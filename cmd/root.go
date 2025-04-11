package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	// Initialize Viper configuration on package init
	cobra.OnInitialize(initConfig)

	// --- Persistent Flags (available to all subcommands) ---
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print verbose test output")

	// --- Local Flags (only for the root command if run without subcommands) ---
	// Add flags that are shared between root default action and testCmd here
	// Or add them specifically in test.go if only for the test subcommand
	addTestExecutionFlags(rootCmd)
}

var (
	cfgFile string // Path to config file (optional, Viper finds by default)
	verbose bool

	rootCmd = &cobra.Command{
		Use:   "gotest [directory]",
		Short: "A helper CLI for running Go tests and benchmarks with extra features.",
		Long: `gotest provides enhanced functionality for running Go tests,
benchmarks, managing history, and profiling.

Specify a directory to run tests within it, or run from the current directory.`,
		// Args validation for the optional directory
		Args:         cobra.MaximumNArgs(1),
		RunE:         runTestCommand,
		SilenceUsage: true,
	}
)

// Helper to add flags common to root and test commands
func addTestExecutionFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("subtest", "s", false, "Select and run a specific subtest interactively")
	cmd.Flags().StringP("name", "n", "", "Run a test by its name (regex match)")
	cmd.Flags().BoolP("rerun", "r", false, "Rerun the last test")
	cmd.Flags().BoolP("debug", "d", false, "Run test in debug mode with delve")
	cmd.Flags().Bool("cover", false, "Run the test with coverage and auto launch the viewer")

	// Used for profiling within tests and benchmarks.
	cmd.PersistentFlags().Bool("cpu", false, "Run the test with a CPU profile")
	cmd.PersistentFlags().Bool("mem", false, "Run the test with a memory profile")

	// Bind flags to Viper keys (optional but recommended)
	viper.BindPFlag("debug", cmd.Flags().Lookup("debug"))
	viper.BindPFlag("cover", cmd.Flags().Lookup("cover"))
	viper.BindPFlag("cpu", cmd.Flags().Lookup("cpu"))
	viper.BindPFlag("mem", cmd.Flags().Lookup("mem"))
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		// Cobra prints errors, so we just exit
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		configDir, err := os.UserConfigDir()
		cobra.CheckErr(err) // Use Cobra's helper for errors

		// Search config in ~/.config/gotest directory with name "config" (without extension).
		configPath := filepath.Join(configDir, "gotest")
		viper.AddConfigPath(configPath)
		viper.SetConfigName("config") // name of config file (without extension)
		// Supported types Viper will look for: json, toml, yaml, yml, properties, props, prop, hcl, tfvars, dotenv, env, ini
		viper.SetConfigType("yaml") // Or "toml", "json", etc. Choose one. YAML is common.
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("GOTEST") // e.g. GOTEST_COLORIZEOUTPUT=true
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// --- Set Viper Defaults ---
	viper.SetDefault("colorizeoutput", true)

	err := viper.ReadInConfig()
	switch {
	case err == nil:
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	default:
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, "Error reading config file:", err)
		}
	}
}
