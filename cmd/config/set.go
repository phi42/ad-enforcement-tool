package config

import (
	"fmt"

	cfg "github.com/phi42/ad-enforcement-tool/internal/config"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration default.",
	Args:  cobra.ExactArgs(2),
	RunE:  setRun,
}

func setRun(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]
	if err := cfg.ValidateKey(key); err != nil {
		return err
	}
	cfgPath, err := cfg.ResolveConfigPath(configFileFlag, globalFlag)
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}
	if err := cfg.SetKey(cfgPath, key, value); err != nil {
		return fmt.Errorf("setting config value: %w", err)
	}
	fmt.Printf("Set %s = %s %s\n", key, value, cfg.ResolveConfigScope(configFileFlag, globalFlag))
	return nil
}
