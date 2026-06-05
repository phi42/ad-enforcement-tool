package dsl

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/phi42/ad-enforcement-tool/internal/parser"
	"github.com/phi42/ad-enforcement-tool/rule"
)

// ParseDSL parses a DSL source string and returns the corresponding rule.Spec.
// It runs both syntactic (ANTLR) and semantic (validateIR) checks; the
// returned error joins all detected problems.
func ParseDSL(src string) (*rule.Spec, error) {
	// Custom block bodies contain arbitrary text the ANTLR lexer cannot
	// tokenise, so we extract them up-front and replace each block with the
	// same number of newlines to keep ANTLR's line numbers accurate.
	cleaned, customRules, err := extractCustomBlocks(src)
	if err != nil {
		return nil, err
	}

	input := antlr.NewInputStream(cleaned)
	lexer := parser.NewADELexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewADEParser(stream)

	errListener := &dslErrorListener{}
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errListener)
	p.RemoveErrorListeners()
	p.AddErrorListener(errListener)

	tree := p.File()
	if len(errListener.errors) > 0 {
		return nil, fmt.Errorf("syntax errors:\n%s", strings.Join(errListener.errors, "\n"))
	}

	visitor := &irVisitor{}
	result := tree.Accept(visitor)
	if visitor.err != nil {
		return nil, visitor.err
	}

	ir, ok := result.(*rule.Spec)
	if !ok {
		return nil, fmt.Errorf("internal error: visitor did not return rule.Spec")
	}
	ir.Rules = append(ir.Rules, customRules...)

	if err := validateIR(ir); err != nil {
		return nil, err
	}
	return ir, nil
}

// ParseFile reads the file at path and parses it as a DSL source.
// The path is included in any error returned for diagnostics.
func ParseFile(path string) (*rule.Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	ir, err := ParseDSL(string(data))
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return ir, nil
}

// customBlockRE matches the header of a custom block:  custom "name" {
// Group 1 captures the rule name.
var customBlockRE = regexp.MustCompile(
	`(?m)custom\s+"((?:[^"\\]|\\.)*)"` + `\s*\{`,
)

// extractCustomBlocks pulls every `custom "<name>" { ... }` block out of src
// and returns:
//   - the source with each block replaced by the same number of newline
//     characters (so ANTLR line numbers stay correct);
//   - one *rule.Rule per extracted block, with IsCustomRule=true and RawBody
//     set to the contents between the braces.
func extractCustomBlocks(src string) (string, []*rule.Rule, error) {
	var rules []*rule.Rule
	cleaned := src

	for {
		loc := customBlockRE.FindStringIndex(cleaned)
		if loc == nil {
			break
		}

		match := customBlockRE.FindStringSubmatch(cleaned[loc[0]:])
		ruleName := match[1]
		line := strings.Count(cleaned[:loc[0]], "\n") + 1

		braceStart := loc[1] - 1 // position of '{'
		closePos := findMatchingBrace(cleaned, braceStart)
		if closePos < 0 {
			return "", nil, fmt.Errorf("line %d: unterminated custom block %q", line, ruleName)
		}

		rawBody := strings.TrimSpace(cleaned[braceStart+1 : closePos])
		rules = append(rules, &rule.Rule{
			Name:         ruleName,
			IsCustomRule: true,
			RawBody:      rawBody,
		})

		blockText := cleaned[loc[0] : closePos+1]
		replacement := strings.Repeat("\n", strings.Count(blockText, "\n"))
		cleaned = cleaned[:loc[0]] + replacement + cleaned[closePos+1:]
	}

	return cleaned, rules, nil
}

// findMatchingBrace returns the index of the '}' that pairs with the '{' at
// openPos, accounting for nested braces. -1 if unmatched.
func findMatchingBrace(src string, openPos int) int {
	depth := 1
	for i := openPos + 1; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// unquote strips the surrounding double quotes produced by the ANTLR STRING
// token and unescapes embedded \" sequences.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	return strings.ReplaceAll(s, `\"`, `"`)
}

// dslErrorListener captures ANTLR syntax errors with line/column.
type dslErrorListener struct {
	antlr.DefaultErrorListener
	errors []string
}

func (l *dslErrorListener) SyntaxError(
	_ antlr.Recognizer, _ interface{},
	line, column int,
	msg string,
	_ antlr.RecognitionException,
) {
	l.errors = append(l.errors, fmt.Sprintf("line %d:%d: %s", line, column, msg))
}
