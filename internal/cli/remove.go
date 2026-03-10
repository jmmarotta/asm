package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

const removeOriginFlag = "origin"

func newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove [<name>...] [--origin <origin>...]",
		Short:   "Remove skills",
		Args:    validateRemoveArgs,
		Aliases: []string{"rm", "uninstall"},
		RunE:    runRemove,
	}

	cmd.Flags().StringSlice(removeOriginFlag, nil, "Origin selector to remove all matching skills")

	return cmd
}

func validateRemoveArgs(cmd *cobra.Command, args []string) error {
	origins, err := cmd.Flags().GetStringSlice(removeOriginFlag)
	if err != nil {
		return err
	}
	if !hasNonEmptyValues(args) && !hasNonEmptyValues(origins) {
		return fmt.Errorf("requires at least one skill name or --origin")
	}
	return nil
}

func hasNonEmptyValues(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func runRemove(cmd *cobra.Command, args []string) error {
	origins, err := cmd.Flags().GetStringSlice(removeOriginFlag)
	if err != nil {
		return err
	}

	report, err := asm.Remove(args, origins)
	if err != nil {
		return err
	}
	printRemoveReport(report, cmd.OutOrStdout(), cmd.ErrOrStderr())
	return nil
}
