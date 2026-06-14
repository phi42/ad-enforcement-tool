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
	cfgPath, err := cfg.ResolveConfigPath(globalFlag)
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}
	if err := cfg.SetKey(cfgPath, key, value); err != nil {
		return fmt.Errorf("setting config value: %w", err)
	}
	fmt.Printf("Set %s = %s %s\n", key, value, cfg.ResolveConfigScope(globalFlag))
	printInputOverrideHints(cfgPath, key)
	return nil
}

// printInputOverrideHints informs the user when a newly set input default
// interacts with another input default (general vs. command-specific).
func printInputOverrideHints(cfgPath, key string) {
	switch key {
	case cfg.DefaultInput:
		if v, ok, _ := cfg.GetKey(cfgPath, cfg.DefaultCompileInput); ok && v != "" {
			fmt.Printf("Note: %s is already set and takes priority over %s for the compile command.\n",
				cfg.DefaultCompileInput, cfg.DefaultInput)
		}
		if v, ok, _ := cfg.GetKey(cfgPath, cfg.DefaultVerifyInput); ok && v != "" {
			fmt.Printf("Note: %s is already set and takes priority over %s for the verify command.\n",
				cfg.DefaultVerifyInput, cfg.DefaultInput)
		}
	case cfg.DefaultCompileInput:
		if v, ok, _ := cfg.GetKey(cfgPath, cfg.DefaultInput); ok && v != "" {
			fmt.Printf("Note: %s overrides the general %s for the compile command.\n",
				key, cfg.DefaultInput)
		}
	case cfg.DefaultVerifyInput:
		if v, ok, _ := cfg.GetKey(cfgPath, cfg.DefaultInput); ok && v != "" {
			fmt.Printf("Note: %s overrides the general %s for the verify command.\n",
				key, cfg.DefaultInput)
		}
	}
}
