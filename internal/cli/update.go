package cli

import (
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update [name]",
		Short:   "Update skill revisions",
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"up"},
		RunE:    runUpdate,
	}

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	report, err := asm.Update(name)
	if err != nil {
		return err
	}
	printUpdateReport(report, cmd.OutOrStdout(), cmd.ErrOrStderr())
	return nil
}
