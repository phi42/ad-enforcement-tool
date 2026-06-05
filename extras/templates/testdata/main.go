// Command gen writes a minimal ADE Spec protobuf message to stdout.
// It mirrors the rules defined in ../sample.rule so you can pipe the result
// directly to any plugin without running the ade tool:
//
//	go run main.go | path/to/plugin
//	go run main.go > sample.bin && ./plugin < sample.bin
package main

import (
	"os"

	"github.com/phi42/ad-enforcement-tool/rule"
	"google.golang.org/protobuf/proto"
)

func main() {
	spec := &rule.Spec{
		Adr: &rule.Adr{
			Id:    "0001",
			Title: "Use Layered Architecture",
		},
		Mode: rule.InvocationMode_MODE_VERIFY,
		Selectors: []*rule.Selector{
			{
				Name:    "Domain",
				Pattern: "com.example.domain",
				Kind:    rule.SelectorKind_SELECTOR_COMPONENT,
			},
			{
				Name:    "App",
				Pattern: "com.example.app",
				Kind:    rule.SelectorKind_SELECTOR_COMPONENT,
			},
		},
		Rules: []*rule.Rule{
			{
				Name:     "no_upward_deps",
				Kind:     rule.RuleKind_RULE_NOT_DEPEND,
				Severity: rule.Severity_SEVERITY_ERROR,
				From: &rule.TargetRef{
					Value: "App",
					Kind:  rule.SelectorKind_SELECTOR_COMPONENT,
				},
				Targets: []*rule.TargetRef{
					{
						Value: "Domain",
						Kind:  rule.SelectorKind_SELECTOR_COMPONENT,
					},
				},
			},
			{
				Name:       "readme_exists",
				Severity:   rule.Severity_SEVERITY_WARNING,
				IsFileRule: true,
				Checks: []*rule.Check{
					{
						Kind: rule.CheckKind_CHECK_FS_MUST_EXIST,
						Path: "README.md",
					},
				},
			},
		},
	}

	data, err := proto.Marshal(spec)
	if err != nil {
		panic(err)
	}
	if _, err := os.Stdout.Write(data); err != nil {
		panic(err)
	}
}
