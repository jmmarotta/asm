package cli

import (
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

func newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install skills to ./skills",
		RunE:  runInstall,
	}

	return cmd
}

func runInstall(cmd *cobra.Command, _ []string) error {
	report, err := asm.Install()
	if err != nil {
		return err
	}
	printInstallReport(report, cmd.OutOrStdout(), cmd.ErrOrStderr())
	return nil
}
