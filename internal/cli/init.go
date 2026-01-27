package cli

import (
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

const initCwdFlag = "cwd"

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a skills manifest",
		RunE:  runInit,
	}

	cmd.Flags().String(initCwdFlag, "", "Initialize in a specific directory")

	return cmd
}

func runInit(cmd *cobra.Command, _ []string) error {
	cwd, err := cmd.Flags().GetString(initCwdFlag)
	if err != nil {
		return err
	}

	report, err := asm.Init(cwd)
	if err != nil {
		return err
	}
	printInitReport(report, cmd.OutOrStdout())
	return nil
}
