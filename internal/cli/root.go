package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
)

const debugFlag = "debug"

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	cobra.OnInitialize(initConfig)

	cmd := &cobra.Command{
		Use:           "asm",
		Short:         "ASM CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			enabled, err := readBoolFlag(cmd, debugFlag)
			if err != nil {
				return err
			}
			debug.Configure(enabled, cmd.ErrOrStderr())
			if enabled {
				debug.Logf("command %s args=%v", cmd.CommandPath(), args)
			}
			return nil
		},
	}

	cmd.PersistentFlags().Bool(debugFlag, false, "Enable debug logging")

	cmd.AddCommand(newLsCommand())
	cmd.AddCommand(newShowCommand())
	cmd.AddCommand(newAddCommand())
	cmd.AddCommand(newUpdateCommand())
	cmd.AddCommand(newRemoveCommand())
	cmd.AddCommand(newInstallCommand())
	cmd.AddCommand(newInitCommand())

	return cmd
}

func initConfig() {
	viper.SetEnvPrefix("ASM")
	viper.AutomaticEnv()
}

func readBoolFlag(cmd *cobra.Command, name string) (bool, error) {
	if cmd.Flags().Lookup(name) != nil {
		return cmd.Flags().GetBool(name)
	}
	if cmd.PersistentFlags().Lookup(name) != nil {
		return cmd.PersistentFlags().GetBool(name)
	}
	if cmd.InheritedFlags().Lookup(name) != nil {
		return cmd.InheritedFlags().GetBool(name)
	}
	return false, fmt.Errorf("flag %q not found", name)
}
