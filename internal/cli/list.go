package cli

import (
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

func newLsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List configured skills",
		RunE:  runLs,
	}

	return cmd
}

func runLs(cmd *cobra.Command, _ []string) error {
	report, err := asm.List()
	if err != nil {
		return err
	}
	return printListReport(report, cmd.OutOrStdout())
}
