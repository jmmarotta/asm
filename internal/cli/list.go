package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newLsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List configured skills",
		RunE:  runLs,
	}

	return cmd
}

func runLs(cmd *cobra.Command, _ []string) error {
	state, err := loadManifest()
	if err != nil {
		return err
	}

	if len(state.Config.Skills) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No skills found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "NAME\tTYPE\tORIGIN\tVERSION\tSUBDIR")
	for _, skill := range state.Config.Skills {
		fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\n",
			skill.Name,
			skill.Type,
			skill.Origin,
			skill.Version,
			skill.Subdir,
		)
	}

	return writer.Flush()
}
