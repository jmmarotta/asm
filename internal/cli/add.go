package cli

import (
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

const addPathFlag = "path"

func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <path-or-url>",
		Short:   "Add a skill",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"a"},
		RunE:    runAdd,
	}

	cmd.Flags().String(addPathFlag, "", "Subdirectory path to install")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	pathFlag, err := cmd.Flags().GetString(addPathFlag)
	if err != nil {
		return err
	}

	report, err := asm.Add(args[0], pathFlag)
	if err != nil {
		return err
	}
	printInstallReport(report, cmd.OutOrStdout(), cmd.ErrOrStderr())
	return nil
}
