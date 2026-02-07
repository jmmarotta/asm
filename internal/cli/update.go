package cli

import (
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

const updatePathFlag = "path"

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update [name|origin]",
		Short:   "Update skill revisions",
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"up"},
		RunE:    runUpdate,
	}

	cmd.Flags().String(updatePathFlag, "", "Subdirectory path used with an origin selector")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	selector := ""
	if len(args) > 0 {
		selector = args[0]
	}

	pathFlag, err := cmd.Flags().GetString(updatePathFlag)
	if err != nil {
		return err
	}

	report, err := asm.Update(selector, pathFlag)
	if err != nil {
		return err
	}
	printUpdateReport(report, cmd.OutOrStdout(), cmd.ErrOrStderr())
	return nil
}
