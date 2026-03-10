package cli

import (
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

func newUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade asm via go install",
		Long:  "Reinstall asm from the canonical Go module path using go install ...@latest.",
		Args:  cobra.NoArgs,
		RunE:  runUpgrade,
	}

	return cmd
}

func runUpgrade(cmd *cobra.Command, _ []string) error {
	report, err := asm.Upgrade()
	if err != nil {
		return err
	}
	printUpgradeReport(report, cmd.OutOrStdout())
	return nil
}
