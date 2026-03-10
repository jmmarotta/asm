package cli

import (
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

const listSortFlag = "sort"

func newLsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List configured skills",
		RunE:  runLs,
	}
	cmd.Flags().String(listSortFlag, string(asm.ListSortName), "Sort rows by: name, origin")

	return cmd
}

func runLs(cmd *cobra.Command, _ []string) error {
	sortValue, err := cmd.Flags().GetString(listSortFlag)
	if err != nil {
		return err
	}

	sortBy, err := asm.ParseListSort(sortValue)
	if err != nil {
		return err
	}

	report, err := asm.List(asm.ListOptions{Sort: sortBy})
	if err != nil {
		return err
	}
	return printListReport(report, cmd.OutOrStdout())
}
