package dslref

import (
	_ "embed"

	"github.com/phi42/ad-enforcement-tool/internal/domain"
)

// Reference contains the full ADE rule DSL language reference.
//
//go:embed dsl.md
var Reference string

// Validate parses content as an ADE rule file and returns any syntax or
// semantic errors. Returns nil if the content is valid.
func Validate(content string) error {
	_, err := domain.ParseDSL(content)
	return err
}
