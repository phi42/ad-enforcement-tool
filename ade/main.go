package main

import (
	"os"

	adecmd "github.com/phi42/ad-enforcement-tool/cmd"
)

func main() {
	cmd := adecmd.NewEnforceCommand()
	cmd.Use = "ade"
	cmd.Short = "Architectural Decision Enforcement Tool (ADE)"
	cmd.Long = `CLI tool for enforcing architectural decisions using rule files.See https://github.com/phi42/ad-enforcement-tool for documentation.`

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
