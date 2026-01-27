package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

func printInstallReport(report asm.InstallReport, out io.Writer, errOut io.Writer) {
	for _, warning := range report.Warnings {
		if warning.Target != "" {
			fmt.Fprintf(errOut, "warning: %s (%s)\n", warning.Message, warning.Target)
			continue
		}
		fmt.Fprintf(errOut, "warning: %s\n", warning.Message)
	}

	if report.NoSkills {
		fmt.Fprintln(out, "No skills found.")
		return
	}

	fmt.Fprintf(out, "Installed: %d, Pruned: %d, Warnings: %d\n", report.Linked, report.Pruned, len(report.Warnings))
}

func printListReport(report asm.ListReport, out io.Writer) error {
	if report.NoSkills {
		fmt.Fprintln(out, "No skills found.")
		return nil
	}

	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "NAME\tTYPE\tORIGIN\tVERSION\tSUBDIR")
	for _, skill := range report.Skills {
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

func printShowReport(report asm.ShowReport, out io.Writer) error {
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(out, string(payload))
	return nil
}

func printInitReport(_ asm.InitReport, out io.Writer) {
	fmt.Fprintln(out, "Initialized skills.jsonc")
}

func printUpdateReport(report asm.UpdateReport, out io.Writer, errOut io.Writer) {
	printInstallReport(report.Install, out, errOut)
	if report.Install.NoSkills {
		return
	}
	if len(report.UpdatedOrigins) > 0 {
		fmt.Fprintf(out, "Updated origins: %s\n", strings.Join(report.UpdatedOrigins, ", "))
	}
}

func printRemoveReport(report asm.RemoveReport, out io.Writer, errOut io.Writer) {
	printInstallReport(report.Install, out, errOut)
	if report.Install.NoSkills {
		return
	}
	if report.Removed.Name != "" {
		if report.Removed.Origin != "" {
			fmt.Fprintf(out, "Removed: %s (%s)\n", report.Removed.Name, report.Removed.Origin)
		} else {
			fmt.Fprintf(out, "Removed: %s\n", report.Removed.Name)
		}
	}
	if report.PrunedStore {
		fmt.Fprintf(out, "Pruned store: %s\n", report.Removed.Origin)
	}
}
