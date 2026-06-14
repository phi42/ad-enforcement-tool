package config

import (
	"fmt"

	cfg "github.com/phi42/ad-enforcement-tool/internal/config"
	"github.com/spf13/cobra"
)

var unsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Remove a configuration default.",
	Args:  cobra.ExactArgs(1),
	RunE:  unsetRun,
}

func unsetRun(cmd *cobra.Command, args []string) error {
	key := args[0]
	if err := cfg.ValidateKey(key); err != nil {
		return err
	}
	cfgPath, err := cfg.ResolveConfigPath(globalFlag)
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}
	if err := cfg.UnsetKey(cfgPath, key); err != nil {
		return fmt.Errorf("unsetting config value: %w", err)
	}
	fmt.Printf("Unset %s %s\n", key, cfg.ResolveConfigScope(globalFlag))
	return nil
}
