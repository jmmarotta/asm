package cli

import (
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

func newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a skill",
		Args:  cobra.ExactArgs(1),
		RunE:  runRemove,
	}

	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	report, err := asm.Remove(args[0])
	if err != nil {
		return err
	}
	printRemoveReport(report, cmd.OutOrStdout(), cmd.ErrOrStderr())
	return nil
}
