package cli

import (
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show a configured skill",
		Args:  cobra.ExactArgs(1),
		RunE:  runShow,
	}
	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	report, err := asm.Show(args[0])
	if err != nil {
		return err
	}
	return printShowReport(report, cmd.OutOrStdout())
}
