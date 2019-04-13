package config

import (
	"flag"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Configurable interface {
	// AddFlags adds CLI flags for configuring this component.
	AddFlags(flagSet *flag.FlagSet)

	// InitFromViper initializes this component with properties from spf13/viper.
	InitFromViper(v *viper.Viper)
}

func AddFlags(v *viper.Viper, command *cobra.Command, addFlagsFns ...func(*flag.FlagSet)) (*viper.Viper, *cobra.Command) {
	flagSet := new(flag.FlagSet)
	for _, addFlags := range addFlagsFns {
		addFlags(flagSet)
	}
	command.Flags().AddGoFlagSet(flagSet)

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.BindPFlags(command.Flags())
	return v, command
}
