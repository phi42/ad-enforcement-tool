package config

import (
	"fmt"

	cfg "github.com/phi42/ad-enforcement-tool/internal/config"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get the effective value of a configuration key.",
	Args:  cobra.ExactArgs(1),
	RunE:  getRun,
}

func getRun(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := cfg.Viper().GetString(key)
	if value == "" {
		return fmt.Errorf("%s is not set", key)
	}
	fmt.Println(value)
	return nil
}
