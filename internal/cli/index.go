package cli

import (
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

const indexOutputFlag = "output"

func newIndexCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Generate a Markdown index of skills",
		RunE:  runIndex,
	}

	cmd.Flags().StringP(indexOutputFlag, "o", "", "Output path for the index Markdown file")

	return cmd
}

func runIndex(cmd *cobra.Command, _ []string) error {
	output, err := cmd.Flags().GetString(indexOutputFlag)
	if err != nil {
		return err
	}

	report, err := asm.Index(output)
	if err != nil {
		return err
	}
	printIndexReport(report, cmd.OutOrStdout(), cmd.ErrOrStderr())
	return nil
}
