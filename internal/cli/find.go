package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

func newFindCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "find <query...>",
		Short:   "Find skills on skills.sh",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"f"},
		RunE:    runFind,
	}

	return cmd
}

func runFind(cmd *cobra.Command, args []string) error {
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return fmt.Errorf("query is required")
	}

	report, err := asm.Find(query)
	if err != nil {
		return err
	}
	printFindReport(report, cmd.OutOrStdout())
	return nil
}
