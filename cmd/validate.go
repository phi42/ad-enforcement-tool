package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/phi42/ad-enforcement-tool/internal/config"
	"github.com/phi42/ad-enforcement-tool/internal/dsl"
	"github.com/phi42/ad-enforcement-tool/internal/runner"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate rule file(s) syntax without executing any plugin.",
	Long: `Validate one or more rule files for syntax errors.

This command parses the rule files and reports any syntax or semantic errors
without actually executing any plugin.

When no -i flag is given, the value of defaults.input in the active config is
used as the input path.

Examples:
  ade validate -i rules/0001.rule
  ade validate -i rules/
  ade validate -i rules/0001.rule -i rules/0002.rule
  ade validate                    # uses defaults.input from config`,
	RunE: validateRun,
}

func init() {
	validateCmd.Flags().StringArrayP("input", "i", []string{},
		"input path(s) containing rule files (can be file or directory; falls back to "+config.DefaultInput+" in config)")
}

// validateRun expands every --input path into its constituent .rule files,
// parses each one, and prints a checkmark or error per file. The command
// exits with a non-zero status if any file fails.
func validateRun(cmd *cobra.Command, args []string) error {
	inputs, err := cmd.Flags().GetStringArray("input")
	if err != nil {
		return fmt.Errorf("reading input flag: %w", err)
	}
	if len(inputs) == 0 {
		fallback := config.Viper().GetString(config.DefaultInput)
		if strings.TrimSpace(fallback) == "" {
			return fmt.Errorf("at least one --input path is required (or set %s in config)", config.DefaultInput)
		}
		inputs = []string{fallback}
	}

	var ruleFiles []string
	for _, p := range inputs {
		files, err := runner.CollectRuleFiles(p)
		if err != nil {
			return fmt.Errorf("collecting files from %q: %w", p, err)
		}
		ruleFiles = append(ruleFiles, files...)
	}
	if len(ruleFiles) == 0 {
		return fmt.Errorf("no rule files found")
	}

	hasErrors := false
	for _, file := range ruleFiles {
		if _, err := dsl.ParseFile(file); err != nil {
			hasErrors = true
			fmt.Fprintf(os.Stderr, "X %s: %v\n", file, err)
		} else {
			fmt.Fprintf(os.Stdout, "✓ %s\n", file)
		}
	}
	if hasErrors {
		return fmt.Errorf("validation failed for one or more files")
	}
	return nil
}
