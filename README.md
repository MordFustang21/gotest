# gotest CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/MordFustang21/gotest)](https://goreportcard.com/report/github.com/MordFustang21/gotest)
<!-- Add other badges here if you have CI/CD, code coverage, etc. -->

A helper CLI for enhancing Go testing and benchmarking workflows with interactive selection, history, profiling, debugging, and more.

## Overview

`gotest` aims to simplify common tasks around running Go tests and benchmarks. It provides features beyond the standard `go test` command, such as:

*   **Interactive Selection:** Easily pick specific tests or benchmarks to run from a list.
*   **Subtest Support:** Detects and allows running specific `t.Run` subtests.
*   **History:** View past test/benchmark runs and re-run them easily.
*   **Debugging:** Seamlessly launch `dlv` (Delve debugger) for a selected test.
*   **Profiling:** Generate CPU and memory profiles with simple flags.
*   **Flame Graphs:** Automatically generate and serve interactive SVG flame graphs for CPU profiles.
*   **Coverage:** Run tests with coverage and automatically open the HTML report.
*   **Configuration:** Customize behavior via a config file or environment variables.

## Features

*   Run Go tests (`go test`)
    *   Run all tests in the current or specified directory.
    *   Interactively select a specific test function (`-s`, `--subtest`).
    *   Run a specific test or subtest by name (regex match) (`-n`, `--name`).
    *   Rerun the last executed test command (`-r`, `--rerun`).
    *   Run tests with verbose output (`-v`, `--verbose`).
    *   Run tests with coverage analysis and viewer (`--cover`).
    *   Debug tests interactively using Delve (`-d`, `--debug`).
*   Run Go benchmarks (`go test -bench`)
    *   Run all benchmarks in the current or specified directory.
    *   Interactively select a specific benchmark function.
*   Profiling (for both tests and benchmarks)
    *   Generate CPU profiles (`--cpu`).
    *   Generate interactive Flame Graphs from CPU profiles.
    *   Generate memory profiles (`--mem`).
*   Command History (`history` subcommand)
    *   View a list of previously executed commands.
    *   Select and re-run a command from history.
*   Colorized output for test results (PASS/FAIL).
*   Configuration via `~/.config/gotest/config.yaml` or environment variables (e.g., `GOTEST_COLORIZEOUTPUT=false`).

## Installation

### Prerequisites

*   **Go:** Version 1.18 or later recommended.
*   **Delve (`dlv`):** Required for the debugging feature (`-d`). Install via `go install github.com/go-delve/delve/cmd/dlv@latest`.
*   **Perl:** Required for generating flame graphs (`--cpu`). Usually pre-installed on Linux and macOS.

### Install `gotest`

```bash
go install github.com/MordFustang21/gotest@latest
```

Ensure your Go bin directory (e.g., `$GOPATH/bin` or `$HOME/go/bin`) is in your system's `PATH`.

## Usage

### General Syntax

```bash
gotest [command] [directory] [flags]
```

*   If `[directory]` is omitted, the current working directory is used.
*   If `[command]` is omitted, it defaults to the `test` command.

### Commands

1.  **`test` (Default)**: Run Go tests.
    ```bash
    # Run all tests in the current directory
    gotest

    # Run all tests in a specific directory
    gotest ./internal/users

    # Interactively select a test/subtest to run
    gotest -s
    # or
    gotest test -s

    # Run a test/subtest matching a name (regex)
    gotest -n TestMySpecificFunction
    gotest -n TestTableDriven/subtest_case_1

    # Rerun the last test command executed in the current module
    gotest -r

    # Run tests with coverage and open the HTML report
    gotest --cover

    # Debug an interactively selected test
    gotest -s -d

    # Run a test with CPU profiling (generates flame graph)
    gotest -s --cpu

    # Run a test with memory profiling
    gotest -s --mem
    ```

2.  **`bench`**: Run Go benchmarks.
    ```bash
    # Interactively select a benchmark to run in the current directory
    gotest bench

    # Interactively select a benchmark in a specific directory
    gotest bench ./pkg/performance

    # Run a selected benchmark with CPU profiling (generates flame graph)
    gotest bench --cpu
    ```

3.  **`history`**: View and re-run past commands.
    ```bash
    # Show history and select a command to re-run
    gotest history
    ```

### Common Flags

*   `-s`, `--subtest`: Interactively select a test/subtest to run (for `test` command).
*   `-n`, `--name string`: Run tests/subtests matching the given name (regex) (for `test` command).
*   `-r`, `--rerun`: Rerun the last command executed in the current module (for `test` command).
*   `-d`, `--debug`: Run the selected test in debug mode with Delve (requires `dlv`).
*   `--cover`: Run tests with coverage and open the HTML report.
*   `--cpu`: Generate a CPU profile. For tests, also generates and serves a flame graph.
*   `--mem`: Generate a memory profile.
*   `-v`, `--verbose`: Enable verbose test output (`go test -v`).

## Configuration

`gotest` uses Viper for configuration. Settings can be provided via:

1.  **Config File:** Create a YAML file at `~/.config/gotest/config.yaml` (path may vary slightly based on OS - uses `os.UserConfigDir`).

    *Example `config.yaml`:*
    ```yaml
    # Enable/disable colorized output (default: true)
    colorizeoutput: true

    # Add other future configuration options here
    ```

2.  **Environment Variables:** Prefix environment variables with `GOTEST_`. Use underscores (`_`) instead of dots (`.`).

    *Example:*
    ```bash
    export GOTEST_COLORIZEOUTPUT=false
    gotest
    ```

### Available Options

*   `colorizeoutput` (boolean, default: `true`): Enable/disable ANSI color codes in the output.

## Examples

```bash
# Run all tests in the current directory verbosely
gotest -v

# Interactively select a test and run it with coverage
gotest -s --cover

# Interactively select a benchmark and profile its CPU usage
gotest bench -s --cpu

# Debug the test named TestUserCreation
gotest -n TestUserCreation -d

# View history and re-run an old command
gotest history

# Rerun the very last command
gotest -r
```

## Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request.

## License
This project is licensed under the [GPLV3] License - see the LICENSE.md file for details.
