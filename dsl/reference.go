// Package dsl exposes the ADE rule DSL language reference as an embedded
// string and provides a lightweight [Validate] helper.
//
// [Reference] embeds dsl-reference.md, the full DSL specification. Other tools
// can embed this package to ship the same language reference without needing
// to copy the file.
package dsl

import (
	_ "embed"

	dsli "github.com/phi42/ad-enforcement-tool/internal/dsl"
)

// Reference contains the full ADE rule DSL language reference.
//
//go:embed dsl-reference.md
var Reference string

// Validate parses content as an ADE rule file and returns any syntax or
// semantic error. It returns nil if the content is valid.
func Validate(content string) error {
	_, err := dsli.ParseDSL(content)
	return err
}
