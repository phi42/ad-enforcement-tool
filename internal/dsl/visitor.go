package dsl

import (
	"fmt"

	"github.com/phi42/ad-enforcement-tool/internal/parser"
	"github.com/phi42/ad-enforcement-tool/rule"
)

// irVisitor walks the ANTLR parse tree and emits a rule.Spec.
//
// All Visit*/visit* methods record the first error encountered in v.err and
// then bail out; callers must check v.err after each top-level call.
type irVisitor struct {
	parser.BaseADEVisitor
	err error
}

// VisitFile is the entry point: it builds the full rule.Spec from a FileContext.
func (v *irVisitor) VisitFile(ctx *parser.FileContext) interface{} {
	ir := &rule.Spec{}
	selectors := make(map[string]bool)
	ruleNames := make(map[string]bool)

	if adrCtx := ctx.AdrDecl(); adrCtx != nil {
		if adr, ok := adrCtx.(*parser.AdrDeclContext); ok {
			adrIR := v.VisitAdrDecl(adr)
			if v.err != nil {
				return nil
			}
			ir.Adr = adrIR.(*rule.Adr)
		}
	}

	for _, selCtx := range ctx.AllSelectorDecl() {
		sel, ok := selCtx.(*parser.SelectorDeclContext)
		if !ok {
			continue
		}
		selIR := v.VisitSelectorDecl(sel)
		if v.err != nil {
			return nil
		}
		selector := selIR.(*rule.Selector)
		if selectors[selector.Name] {
			v.err = fmt.Errorf("line %d: duplicate selector name %q", sel.GetStart().GetLine(), selector.Name)
			return nil
		}
		selectors[selector.Name] = true
		ir.Selectors = append(ir.Selectors, selector)
	}

	for _, ruleCtx := range ctx.AllRuleDecl() {
		ruleC, ok := ruleCtx.(*parser.RuleDeclContext)
		if !ok {
			continue
		}
		ruleIR := v.VisitRuleDecl(ruleC)
		if v.err != nil {
			return nil
		}
		r := ruleIR.(*rule.Rule)
		if ruleNames[r.Name] {
			v.err = fmt.Errorf("line %d: duplicate rule name %q", ruleC.GetStart().GetLine(), r.Name)
			return nil
		}
		ruleNames[r.Name] = true
		ir.Rules = append(ir.Rules, r)
	}

	return ir
}

// VisitAdrDecl extracts the ADR id and title.
func (v *irVisitor) VisitAdrDecl(ctx *parser.AdrDeclContext) interface{} {
	strs := ctx.AllSTRING()
	if len(strs) < 2 {
		v.err = fmt.Errorf("line %d: adr requires id and title", ctx.GetStart().GetLine())
		return nil
	}
	return &rule.Adr{
		Id:    unquote(strs[0].GetText()),
		Title: unquote(strs[1].GetText()),
	}
}

// VisitSelectorDecl extracts the selector name, kind, and pattern.
func (v *irVisitor) VisitSelectorDecl(ctx *parser.SelectorDeclContext) interface{} {
	strs := ctx.AllSTRING()
	if len(strs) < 2 {
		v.err = fmt.Errorf("line %d: selector requires name and pattern", ctx.GetStart().GetLine())
		return nil
	}

	var kind rule.SelectorKind
	switch {
	case ctx.COMPONENT() != nil:
		kind = rule.SelectorKind_SELECTOR_COMPONENT
	case ctx.CLASS() != nil:
		kind = rule.SelectorKind_SELECTOR_CLASS
	case ctx.INTERFACE() != nil:
		kind = rule.SelectorKind_SELECTOR_INTERFACE
	case ctx.PATH() != nil:
		kind = rule.SelectorKind_SELECTOR_PATH
	}

	return &rule.Selector{
		Name:    unquote(strs[0].GetText()),
		Kind:    kind,
		Pattern: unquote(strs[1].GetText()),
	}
}

// VisitRuleDecl builds a single rule.Rule from a RuleDeclContext, walking all
// of its statements (assertions, exclusions, severity).
func (v *irVisitor) VisitRuleDecl(ctx *parser.RuleDeclContext) interface{} {
	ir := &rule.Rule{
		Name: unquote(ctx.STRING().GetText()),
	}

	if ruleTypeCtx := ctx.RuleType(); ruleTypeCtx != nil {
		if rt, ok := ruleTypeCtx.(*parser.RuleTypeContext); ok {
			ir.IsFileRule = rt.FILE() != nil
		}
	}

	for _, stmtCtx := range ctx.AllRuleStmt() {
		if assertCtx := stmtCtx.AssertionStmt(); assertCtx != nil {
			if assert, ok := assertCtx.(*parser.AssertionStmtContext); ok {
				v.visitAssertionStmt(assert, ir)
				if v.err != nil {
					return nil
				}
			}
		}
		if exclCtx := stmtCtx.ExcludeStmt(); exclCtx != nil {
			excl := v.visitExcludeStmt(exclCtx)
			if v.err != nil {
				return nil
			}
			if excl != nil {
				ir.Excludes = append(ir.Excludes, excl)
			}
		}
		if sevCtx := stmtCtx.SeverityStmt(); sevCtx != nil {
			if sev, ok := sevCtx.(*parser.SeverityStmtContext); ok {
				ir.Severity = v.visitSeverityStmt(sev)
			}
		}
	}

	return ir
}

// visitAssertionStmt resolves the subject, modality, and verb phrase of an
// assertion and writes them into ir.
func (v *irVisitor) visitAssertionStmt(ctx *parser.AssertionStmtContext, ir *rule.Rule) {
	ir.From = v.visitSubjectExpr(ctx.SubjectExpr())
	if v.err != nil {
		return
	}
	mod := v.visitMustExpr(ctx.MustExpr())
	v.visitVerbPhrase(ctx.VerbPhrase(), mod, ir)
}

// visitSubjectExpr converts the LHS of an assertion (the "subject") into a
// rule.TargetRef.
func (v *irVisitor) visitSubjectExpr(ctx parser.ISubjectExprContext) *rule.TargetRef {
	if ctx == nil {
		return nil
	}
	switch c := ctx.(type) {
	case *parser.SelectorRefContext:
		return &rule.TargetRef{
			Value:    c.IDENTIFIER().GetText(),
			IsInline: false,
		}
	case *parser.InlineLiteralContext:
		return &rule.TargetRef{
			Value:    unquote(c.STRING().GetText()),
			IsInline: true,
			Kind:     v.getSelectorKind(c.SelectorType()),
		}
	case *parser.InlineMatchContext:
		return &rule.TargetRef{
			Value:    unquote(c.STRING().GetText()),
			IsInline: true,
			Kind:     v.getSelectorKind(c.SelectorType()),
			IsMatch:  true,
		}
	case *parser.InlineTypeContext:
		return &rule.TargetRef{
			IsInline: true,
			Kind:     v.getSelectorKind(c.SelectorType()),
		}
	case *parser.SubsetAllContext:
		return &rule.TargetRef{
			IsInline: true,
			Kind:     v.getSelectorKind(c.SelectorType()),
			Scope:    v.visitTargetExpr(c.TargetExpr()),
		}
	case *parser.SubsetLiteralContext:
		return &rule.TargetRef{
			Value:    unquote(c.STRING().GetText()),
			IsInline: true,
			Kind:     v.getSelectorKind(c.SelectorType()),
			Scope:    v.visitTargetExpr(c.TargetExpr()),
		}
	case *parser.SubsetMatchContext:
		return &rule.TargetRef{
			Value:    unquote(c.STRING().GetText()),
			IsInline: true,
			Kind:     v.getSelectorKind(c.SelectorType()),
			IsMatch:  true,
			Scope:    v.visitTargetExpr(c.TargetExpr()),
		}
	}
	return nil
}

// visitTargetExpr converts the RHS of an assertion (a "target") into a
// rule.TargetRef.
func (v *irVisitor) visitTargetExpr(ctx parser.ITargetExprContext) *rule.TargetRef {
	if ctx == nil {
		return nil
	}
	switch c := ctx.(type) {
	case *parser.TargetSelectorRefContext:
		return &rule.TargetRef{
			Value:    c.IDENTIFIER().GetText(),
			IsInline: false,
		}
	case *parser.TargetInlineLiteralContext:
		return &rule.TargetRef{
			Value:    unquote(c.STRING().GetText()),
			IsInline: true,
			Kind:     v.getSelectorKind(c.SelectorType()),
		}
	case *parser.TargetInlineMatchContext:
		return &rule.TargetRef{
			Value:    unquote(c.STRING().GetText()),
			IsInline: true,
			Kind:     v.getSelectorKind(c.SelectorType()),
			IsMatch:  true,
		}
	case *parser.TargetStringLiteralContext:
		return &rule.TargetRef{
			Value:    unquote(c.STRING().GetText()),
			IsInline: true,
		}
	}
	return nil
}

// getSelectorKind maps a grammar selectorType node to the corresponding
// rule.SelectorKind.
func (v *irVisitor) getSelectorKind(ctx parser.ISelectorTypeContext) rule.SelectorKind {
	if ctx == nil {
		return rule.SelectorKind_SELECTOR_UNSPECIFIED
	}
	switch {
	case ctx.COMPONENT() != nil:
		return rule.SelectorKind_SELECTOR_COMPONENT
	case ctx.CLASS() != nil:
		return rule.SelectorKind_SELECTOR_CLASS
	case ctx.INTERFACE() != nil:
		return rule.SelectorKind_SELECTOR_INTERFACE
	case ctx.PATH() != nil:
		return rule.SelectorKind_SELECTOR_PATH
	}
	return rule.SelectorKind_SELECTOR_UNSPECIFIED
}

// modality is the resolved meaning of a "must"/"must not"/"must only" phrase.
type modality int

const (
	modalityMust modality = iota
	modalityMustNot
	modalityMustOnly
)

func (v *irVisitor) visitMustExpr(ctx parser.IMustExprContext) modality {
	if ctx == nil {
		return modalityMust
	}
	mustCtx, ok := ctx.(*parser.MustExprContext)
	if !ok {
		return modalityMust
	}
	if mustCtx.NOT() != nil {
		return modalityMustNot
	}
	if mustCtx.ONLY() != nil {
		return modalityMustOnly
	}
	return modalityMust
}

// visitVerbPhrase dispatches on the verb phrase variant and updates the rule
// (kind, targets, checks, etc.) accordingly.
func (v *irVisitor) visitVerbPhrase(ctx parser.IVerbPhraseContext, mod modality, ir *rule.Rule) {
	if ctx == nil {
		return
	}
	switch c := ctx.(type) {
	case *parser.DependOnPhraseContext:
		v.applyDependOn(c, mod, ir)
	case *parser.ExistPhraseContext:
		v.applyExist(mod, ir)
	case *parser.ContainPhraseContext:
		v.applyContain(c, mod, ir)
	case *parser.ImplementPhraseContext:
		v.applyImplement(c, mod, ir)
	case *parser.ExtendPhraseContext:
		v.applyExtend(c, mod, ir)
	case *parser.AnnotatedPhraseContext:
		v.applyAnnotated(c, mod, ir)
	case *parser.AccessedByPhraseContext:
		v.applyAccessedBy(c, mod, ir)
	case *parser.AcyclicPhraseContext:
		v.applyAcyclic(mod, ir)
	case *parser.InPhraseContext:
		v.applyIn(c, mod, ir)
	case *parser.MatchPhraseContext:
		v.applyMatch(c, mod, ir)
	case *parser.VisibilityPhraseContext:
		v.applyVisibility(c, mod, ir)
	case *parser.TypeConstraintPhraseContext:
		v.applyTypeConstraint(c, mod, ir)
	}
}

func (v *irVisitor) applyDependOn(c *parser.DependOnPhraseContext, mod modality, ir *rule.Rule) {
	for _, tctx := range c.AllTargetExpr() {
		if t := v.visitTargetExpr(tctx); t != nil {
			ir.Targets = append(ir.Targets, t)
		}
	}
	switch mod {
	case modalityMustNot:
		ir.Kind = rule.RuleKind_RULE_NOT_DEPEND
	case modalityMustOnly:
		ir.Kind = rule.RuleKind_RULE_DEPEND_ONLY
	case modalityMust:
		v.err = fmt.Errorf("invalid dependency rule: use 'must not' or 'must only', not plain 'must'")
	}
}

func (v *irVisitor) applyExist(mod modality, ir *rule.Rule) {
	if mod == modalityMustOnly {
		v.err = fmt.Errorf("invalid 'must only exist' - use 'must exist'")
		return
	}
	kind := rule.CheckKind_CHECK_FS_MUST_EXIST
	if mod == modalityMustNot {
		kind = rule.CheckKind_CHECK_FS_MUST_NOT_EXIST
	}
	ir.Checks = append(ir.Checks, &rule.Check{
		Kind: kind,
		Path: ir.From.Value,
	})
}

func (v *irVisitor) applyContain(c *parser.ContainPhraseContext, mod modality, ir *rule.Rule) {
	if mod == modalityMustOnly {
		v.err = fmt.Errorf("invalid 'must only contain' - use 'must contain'")
		return
	}
	kind := rule.CheckKind_CHECK_FS_MUST_CONTAIN
	if mod == modalityMustNot {
		kind = rule.CheckKind_CHECK_FS_MUST_NOT_CONTAIN
	}
	ir.Checks = append(ir.Checks, &rule.Check{
		Kind:    kind,
		Path:    ir.From.Value,
		Pattern: unquote(c.STRING().GetText()),
	})
}

func (v *irVisitor) applyImplement(c *parser.ImplementPhraseContext, mod modality, ir *rule.Rule) {
	if t := v.visitTargetExpr(c.TargetExpr()); t != nil {
		ir.Targets = append(ir.Targets, t)
	}
	switch mod {
	case modalityMust, modalityMustOnly:
		ir.Kind = rule.RuleKind_RULE_IMPLEMENT
	case modalityMustNot:
		ir.Kind = rule.RuleKind_RULE_NOT_IMPLEMENT
	}
}

func (v *irVisitor) applyExtend(c *parser.ExtendPhraseContext, mod modality, ir *rule.Rule) {
	if t := v.visitTargetExpr(c.TargetExpr()); t != nil {
		ir.Targets = append(ir.Targets, t)
	}
	switch mod {
	case modalityMust, modalityMustOnly:
		ir.Kind = rule.RuleKind_RULE_EXTEND
	case modalityMustNot:
		ir.Kind = rule.RuleKind_RULE_NOT_EXTEND
	}
}

func (v *irVisitor) applyAnnotated(c *parser.AnnotatedPhraseContext, mod modality, ir *rule.Rule) {
	ir.Targets = append(ir.Targets, &rule.TargetRef{
		Value:    unquote(c.STRING().GetText()),
		IsInline: true,
	})
	switch mod {
	case modalityMust:
		ir.Kind = rule.RuleKind_RULE_ANNOTATE
	case modalityMustNot:
		ir.Kind = rule.RuleKind_RULE_NOT_ANNOTATE
	default:
		v.err = fmt.Errorf("invalid annotation rule: use 'must be annotated with' or 'must not be annotated with'")
	}
}

func (v *irVisitor) applyAccessedBy(c *parser.AccessedByPhraseContext, mod modality, ir *rule.Rule) {
	for _, tctx := range c.AllTargetExpr() {
		if t := v.visitTargetExpr(tctx); t != nil {
			ir.Targets = append(ir.Targets, t)
		}
	}
	if mod != modalityMustOnly {
		v.err = fmt.Errorf("accessed by rule requires 'must only' modality")
		return
	}
	ir.Kind = rule.RuleKind_RULE_ACCESSED_BY
}

func (v *irVisitor) applyAcyclic(mod modality, ir *rule.Rule) {
	if mod != modalityMust {
		v.err = fmt.Errorf("acyclic rule requires 'must' modality (not 'must not' or 'must only')")
		return
	}
	ir.Kind = rule.RuleKind_RULE_ACYCLIC
}

func (v *irVisitor) applyIn(c *parser.InPhraseContext, mod modality, ir *rule.Rule) {
	if t := v.visitTargetExpr(c.TargetExpr()); t != nil {
		ir.Targets = append(ir.Targets, t)
	}
	switch mod {
	case modalityMust:
		ir.Kind = rule.RuleKind_RULE_IN
	case modalityMustNot:
		ir.Kind = rule.RuleKind_RULE_NOT_IN
	default:
		v.err = fmt.Errorf("location rule: use 'must be in' or 'must not be in'")
	}
}

func (v *irVisitor) applyMatch(c *parser.MatchPhraseContext, mod modality, ir *rule.Rule) {
	ir.Targets = append(ir.Targets, &rule.TargetRef{
		Value:    unquote(c.STRING().GetText()),
		IsInline: true,
	})
	switch mod {
	case modalityMust:
		ir.Kind = rule.RuleKind_RULE_MATCH
	case modalityMustNot:
		ir.Kind = rule.RuleKind_RULE_NOT_MATCH
	default:
		v.err = fmt.Errorf("naming pattern rule: use 'must match' or 'must not match'")
	}
}

func (v *irVisitor) applyVisibility(c *parser.VisibilityPhraseContext, mod modality, ir *rule.Rule) {
	if mod != modalityMust {
		v.err = fmt.Errorf("Visibility rule requires 'must' modality")
		return
	}
	ir.Kind = rule.RuleKind_RULE_VISIBILITY
	visCtx, ok := c.Visibility().(*parser.VisibilityContext)
	if !ok {
		return
	}
	switch {
	case visCtx.PUBLIC() != nil:
		ir.Visibility = rule.Visibility_VISIBILITY_PUBLIC
	case visCtx.INTERNAL() != nil:
		ir.Visibility = rule.Visibility_VISIBILITY_INTERNAL
	case visCtx.PRIVATE() != nil:
		ir.Visibility = rule.Visibility_VISIBILITY_PRIVATE
	}
}

func (v *irVisitor) applyTypeConstraint(c *parser.TypeConstraintPhraseContext, mod modality, ir *rule.Rule) {
	if mod != modalityMust {
		v.err = fmt.Errorf("type constraint rule requires 'must' modality")
		return
	}
	ir.Kind = rule.RuleKind_RULE_TYPE_CONSTRAINT
	tcCtx, ok := c.TypeConstraint().(*parser.TypeConstraintContext)
	if !ok {
		return
	}
	switch {
	case tcCtx.ABSTRACT() != nil:
		ir.TypeConstraint = rule.TypeConstraint_TYPE_CONSTRAINT_ABSTRACT
	case tcCtx.SEALED() != nil:
		ir.TypeConstraint = rule.TypeConstraint_TYPE_CONSTRAINT_SEALED
	case tcCtx.STATIC() != nil:
		ir.TypeConstraint = rule.TypeConstraint_TYPE_CONSTRAINT_STATIC
	}
}

// visitExcludeStmt converts an `exclude ...` statement into a rule.Exclusion.
func (v *irVisitor) visitExcludeStmt(ctx parser.IExcludeStmtContext) *rule.Exclusion {
	if ctx == nil {
		return nil
	}
	switch c := ctx.(type) {
	case *parser.ExcludeClassContext:
		return &rule.Exclusion{
			Kind:  rule.ExcludeKind_EXCLUDE_CLASS,
			Value: unquote(c.STRING().GetText()),
		}
	case *parser.ExcludeClassImplementingContext:
		return &rule.Exclusion{
			Kind:  rule.ExcludeKind_EXCLUDE_IMPLEMENT_INTERFACE,
			Value: unquote(c.STRING().GetText()),
		}
	case *parser.ExcludeComponentContext:
		return &rule.Exclusion{
			Kind:  rule.ExcludeKind_EXCLUDE_COMPONENT,
			Value: unquote(c.STRING().GetText()),
		}
	case *parser.ExcludePatternContext:
		return &rule.Exclusion{
			Kind:  rule.ExcludeKind_EXCLUDE_CLASS,
			Value: unquote(c.STRING().GetText()),
		}
	}
	return nil
}

// visitSeverityStmt extracts the severity level.
func (v *irVisitor) visitSeverityStmt(ctx *parser.SeverityStmtContext) rule.Severity {
	if ctx == nil {
		return rule.Severity_SEVERITY_UNSPECIFIED
	}
	svCtx := ctx.SeverityValue()
	if svCtx == nil {
		return rule.Severity_SEVERITY_UNSPECIFIED
	}
	switch {
	case svCtx.ERROR() != nil:
		return rule.Severity_SEVERITY_ERROR
	case svCtx.WARNING() != nil:
		return rule.Severity_SEVERITY_WARNING
	}
	return rule.Severity_SEVERITY_UNSPECIFIED
}
