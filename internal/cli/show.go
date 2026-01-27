package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
)

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show a configured skill",
		Args:  cobra.ExactArgs(1),
		RunE:  runShow,
	}
	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	state, err := loadManifest()
	if err != nil {
		return err
	}

	name := args[0]
	skill, found := config.FindSkill(state.Config.Skills, name)
	if !found {
		return fmt.Errorf("skill %q not found", name)
	}

	output := struct {
		config.Skill
		Replace string `json:"replace,omitempty"`
	}{
		Skill:   skill,
		Replace: state.Config.Replace[skill.Origin],
	}

	payload, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(payload))
	return nil
}
