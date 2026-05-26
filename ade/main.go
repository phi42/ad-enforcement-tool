package main

import (
	"os"

	"github.com/phi42/ad-enforcement-tool/enforce"
)

func main() {
	cmd := enforce.NewEnforceCommand()
	cmd.Use = "ade"
	cmd.Short = "Architectural Decision Enforcement Tool (ADE)"
	cmd.Long = "CLI tool for enforcing architectural decisions using rule files.\n\nSee https://github.com/phi42/ad-enforcement-tool for documentation."
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
