package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
)

const findResultsLimit = 6

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
	fmt.Fprintln(writer, "NAME\tORIGIN\tVERSION\tSUBDIR")
	for _, skill := range report.Skills {
		fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\n",
			skill.Name,
			skill.Origin,
			skill.Version,
			skill.Subdir,
		)
	}

	return writer.Flush()
}

func printFindReport(report asm.FindReport, out io.Writer) {
	if len(report.Skills) == 0 {
		fmt.Fprintf(out, "No skills found for %q.\n", report.Query)
		return
	}

	fmt.Fprintln(out, "Install with asm add https://github.com/<owner>/<repo> --path skills/<skill>")
	fmt.Fprintln(out)

	skills := report.Skills
	if len(skills) > findResultsLimit {
		skills = skills[:findResultsLimit]
	}

	for _, skill := range skills {
		skillID := skill.SkillID
		if skillID == "" {
			skillID = skill.Name
		}
		if skill.Source == "" || skillID == "" {
			continue
		}
		fmt.Fprintf(out, "asm add https://github.com/%s --path skills/%s\n", skill.Source, skillID)
		fmt.Fprintf(out, "  https://skills.sh/%s/%s\n", skill.Source, skillID)
		fmt.Fprintln(out)
	}
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
	for _, warning := range report.Warnings {
		fmt.Fprintf(errOut, "warning: %s\n", warning)
	}
	if report.NoChanges {
		fmt.Fprintln(out, "No matching skills removed.")
		return
	}
	printInstallReport(report.Install, out, errOut)
	for _, removed := range report.Removed {
		if removed.Name == "" {
			continue
		}
		if removed.Origin != "" {
			fmt.Fprintf(out, "Removed: %s (%s)\n", removed.Name, removed.Origin)
		} else {
			fmt.Fprintf(out, "Removed: %s\n", removed.Name)
		}
	}
	for _, origin := range report.PrunedStores {
		fmt.Fprintf(out, "Pruned store: %s\n", origin)
	}
}
