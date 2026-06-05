package dsl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/phi42/ad-enforcement-tool/rule"
)

// selectorRefRE matches strings shaped like the grammar's IDENTIFIER token
// ([A-Z][a-zA-Z0-9_]*). Such a token can only have come from a SelectorRef in
// the parse tree, so it must resolve to a declared selector.
var selectorRefRE = regexp.MustCompile(`^[A-Z][a-zA-Z0-9_]*$`)

// validateIR runs semantic checks over a fully-built rule.Spec.
//
// It enforces, in order:
//   - the file declares an `adr ...` block;
//   - rule names are unique (across regular and custom rules);
//   - each rule mixes only file-level or only code-level assertions;
//   - every selector reference resolves to a declared selector, including in
//     scope and target positions.
func validateIR(ir *rule.Spec) error {
	if ir.Adr == nil {
		return fmt.Errorf("missing ADR declaration")
	}

	selectors := make(map[string]bool, len(ir.Selectors))
	for _, sel := range ir.Selectors {
		selectors[sel.Name] = true
	}

	ruleNames := make(map[string]bool, len(ir.Rules))
	for _, r := range ir.Rules {
		if r.IsCustomRule {
			if r.Name == "" {
				return fmt.Errorf("custom rule: missing name")
			}
			if ruleNames[r.Name] {
				return fmt.Errorf("duplicate rule name %q", r.Name)
			}
			ruleNames[r.Name] = true
			continue
		}

		if err := validateAssertionShape(r); err != nil {
			return err
		}
		if err := validateSelectorRefs(r, selectors); err != nil {
			return err
		}
		ruleNames[r.Name] = true
	}

	return nil
}

// validateAssertionShape rejects rules that mix file-level checks (exist,
// contain) with code-level assertions (depend on, implement, ...).
func validateAssertionShape(r *rule.Rule) error {
	hasChecks := len(r.Checks) > 0
	hasCodeAssertion := r.Kind != rule.RuleKind_RULE_UNSPECIFIED && !hasChecks

	if r.IsFileRule && hasCodeAssertion {
		return fmt.Errorf("rule %q: code assertions (depend on, implement, etc.) cannot be used in file rules", r.Name)
	}
	if !r.IsFileRule && hasChecks {
		return fmt.Errorf("rule %q: file system assertions (exist, contain) cannot be used in code rules", r.Name)
	}
	return nil
}

// codeRulesWithSubject is the set of rule kinds whose subject (rule.From)
// must be present and, when not inline, must reference a declared selector.
var codeRulesWithSubject = map[rule.RuleKind]struct{}{
	rule.RuleKind_RULE_DEPEND_ONLY:     {},
	rule.RuleKind_RULE_NOT_DEPEND:      {},
	rule.RuleKind_RULE_ANNOTATE:        {},
	rule.RuleKind_RULE_NOT_ANNOTATE:    {},
	rule.RuleKind_RULE_EXTEND:          {},
	rule.RuleKind_RULE_NOT_EXTEND:      {},
	rule.RuleKind_RULE_IMPLEMENT:       {},
	rule.RuleKind_RULE_NOT_IMPLEMENT:   {},
	rule.RuleKind_RULE_ACCESSED_BY:     {},
	rule.RuleKind_RULE_ACYCLIC:         {},
	rule.RuleKind_RULE_IN:              {},
	rule.RuleKind_RULE_NOT_IN:          {},
	rule.RuleKind_RULE_MATCH:           {},
	rule.RuleKind_RULE_NOT_MATCH:       {},
	rule.RuleKind_RULE_VISIBILITY:      {},
	rule.RuleKind_RULE_TYPE_CONSTRAINT: {},
}

// validateSelectorRefs ensures every named selector reference in r resolves
// to a declared selector.
func validateSelectorRefs(r *rule.Rule, selectors map[string]bool) error {
	if _, ok := codeRulesWithSubject[r.Kind]; ok {
		if r.From == nil {
			return fmt.Errorf("rule %q: missing subject", r.Name)
		}
		if !r.From.IsInline && !selectors[r.From.Value] {
			return fmt.Errorf("rule %q: unknown selector %q", r.Name, r.From.Value)
		}
		if r.From.Scope != nil && !r.From.Scope.IsInline && !selectors[r.From.Scope.Value] {
			return fmt.Errorf("rule %q: unknown selector %q in scope", r.Name, r.From.Scope.Value)
		}
		for _, t := range r.Targets {
			if !t.IsInline && !selectors[t.Value] {
				return fmt.Errorf("rule %q: unknown selector %q", r.Name, t.Value)
			}
		}
	}

	if r.IsFileRule {
		for _, c := range r.Checks {
			if !selectorRefRE.MatchString(c.Path) || selectors[c.Path] {
				continue
			}
			// A path-shaped IDENTIFIER without a declared selector. Hint at
			// the lowercase-first variant if one was declared, since
			// IDENTIFIER tokens require an uppercase first letter.
			lower := strings.ToLower(c.Path[:1]) + c.Path[1:]
			if selectors[lower] {
				return fmt.Errorf(
					"rule %q: undefined selector %q (selector %q is defined but selector names must start with an uppercase letter to be referenced by name)",
					r.Name, c.Path, lower,
				)
			}
			return fmt.Errorf("rule %q: undefined selector %q", r.Name, c.Path)
		}
	}

	return nil
}
