package perfchecklint

import (
	"fmt"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
)

func report(pass *analysis.Pass, pos token.Pos, rule ruleset.Rule, detail string) {
	pass.Report(analysis.Diagnostic{
		Pos:      pos,
		Message:  formatMessage(rule, detail),
		Category: rule.Category,
	})
}

func formatMessage(rule ruleset.Rule, detail string) string {
	detail = normalizeSentence(detail)
	summary := normalizeSentence(rule.ProblemSummary)
	hint := normalizeSentence(rule.FixHint)
	return fmt.Sprintf("[%s] %s Why: %s Fix: %s", rule.ID, detail, summary, hint)
}

func normalizeSentence(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return text
	}
	if strings.HasSuffix(text, ".") || strings.HasSuffix(text, "!") || strings.HasSuffix(text, "?") {
		return text
	}
	return text + "."
}
