package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	cobra.OnInitialize(initConfig)

	cmd := &cobra.Command{
		Use:          "asm",
		Short:        "ASM CLI",
		SilenceUsage: true,
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newShowCommand())
	cmd.AddCommand(newAddCommand())
	cmd.AddCommand(newUpdateCommand())
	cmd.AddCommand(newRemoveCommand())
	cmd.AddCommand(newSyncCommand())
	cmd.AddCommand(newTargetCommand())

	return cmd
}

func initConfig() {
	viper.SetEnvPrefix("ASM")
	viper.AutomaticEnv()
}
