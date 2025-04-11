package cmd

import (
	"errors"
	"fmt"

	"github.com/MordFustang21/gotest/pkg/history" // Assuming history logic is here
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(historyCmd)
	// No specific flags for this command usually
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Select and re-run a command from history",
	Long:  `Displays past executed test/benchmark commands and allows selecting one to re-run.`,
	Args:  cobra.NoArgs, // Does not take directory/test args directly
	RunE: func(cmd *cobra.Command, args []string) error {
		selectedEntry, err := history.SelectHistory()
		switch {
		case err == nil:
		case errors.Is(err, promptui.ErrInterrupt):
			fmt.Println("History selection cancelled.")
			return nil
		default:
			return fmt.Errorf("error selecting from history: %w", err)
		}

		if selectedEntry == nil {
			// Should be handled by SelectHistory, but defensively check
			fmt.Println("No history entry selected.")
			return nil
		}

		fmt.Println("Re-running command from history:")
		err = history.RunHistoryEntry(*selectedEntry)
		if err != nil {
			return fmt.Errorf("failed to re-run command from history: %w", err)
		}

		return nil
	},
}
