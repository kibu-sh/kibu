package kibuenum

import (
	"encoding/json"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rogpeppe/go-internal/testscript"
	"golang.org/x/tools/go/analysis"
)

// writeScriptResults renders the target-neutral spec and source diagnostics into
// stable text for fixture comparisons. It does not influence analyzer behavior.
func writeScriptResults(ts *testscript.TestScript, root string, passes []*analysis.Pass, reported []analysis.Diagnostic) {
	sort.Slice(passes, func(i, j int) bool { return passes[i].Pkg.Path() < passes[j].Pkg.Path() })
	formatter := scriptFormatter{ts: ts, root: root}
	var summary, locations strings.Builder
	var specs []*Result
	for _, pass := range passes {
		result, ok := FromPass(pass)
		if !ok {
			ts.Fatalf("missing kibuenum result for %s", pass.Pkg.Path())
		}
		specs = append(specs, result)
		resultSummary, resultLocations := formatter.formatResult(result)
		summary.WriteString(resultSummary)
		locations.WriteString(resultLocations)
	}

	data, err := json.MarshalIndent(specs, "", "  ")
	ts.Check(err)
	// pipeline.Run uses a shared FileSet for all packages in this load.
	diagnostics := formatter.formatDiagnostics(passes[0].Fset, reported)
	for name, contents := range map[string]string{
		"result.txt":      summary.String(),
		"locations.txt":   locations.String(),
		"result.json":     strings.ReplaceAll(string(data), root, "$SRC") + "\n",
		"diagnostics.txt": diagnostics,
	} {
		ts.Check(os.WriteFile(ts.MkAbs(name), []byte(contents), 0644))
	}
}

// scriptFormatter shares source-path normalization across the text formats.
type scriptFormatter struct {
	ts   *testscript.TestScript
	root string
}

func (f scriptFormatter) formatResult(result *Result) (string, string) {
	var summary, locations strings.Builder
	for _, declaration := range result.Enums {
		enumSummary, enumLocations := f.formatEnum(declaration)
		summary.WriteString(enumSummary)
		locations.WriteString(enumLocations)
	}
	return summary.String(), locations.String()
}

func (f scriptFormatter) formatEnum(declaration Enum) (string, string) {
	var summary, locations strings.Builder
	summary.WriteString(formatEnumSummary(declaration))
	locations.WriteString(f.formatEnumLocations(declaration))
	for _, member := range declaration.Members {
		summary.WriteString(formatMemberSummary(member))
		locations.WriteString(f.formatMemberLocations(member))
	}
	return summary.String(), locations.String()
}

func formatEnumSummary(declaration Enum) string {
	return fmt.Sprintf("enum %s.%s type=%s.%s underlying=%s\n",
		declaration.Declaration.PackagePath, declaration.Declaration.Name,
		declaration.Type.PackagePath, declaration.Type.Name, declaration.Underlying)
}

func formatMemberSummary(member Member) string {
	return fmt.Sprintf("  %s.%s value=%q label=%q description=%q\n",
		member.Constant.PackagePath, member.Constant.Name,
		member.WireValue, member.Label, member.Description)
}

func (f scriptFormatter) formatEnumLocations(declaration Enum) string {
	return fmt.Sprintf("%s declaration=%s type=%s call=%s\n",
		declaration.Declaration.Name, f.position(declaration.Declaration.Position),
		f.position(declaration.Type.Position), f.span(declaration.Source))
}

func (f scriptFormatter) formatMemberLocations(member Member) string {
	return fmt.Sprintf("  %s constant=%s member=%s value=%s label=%s description=%s\n",
		member.Constant.Name, f.position(member.Constant.Position), f.span(member.Source),
		f.span(member.ValueSource), f.span(member.LabelSource), f.span(member.DescriptionSource))
}

func (f scriptFormatter) formatDiagnostics(files *token.FileSet, reported []analysis.Diagnostic) string {
	var lines []string
	for _, diagnostic := range reported {
		lines = append(lines, f.formatDiagnostic(files, diagnostic))
	}
	sort.Strings(lines)
	return strings.Join(lines, "")
}

func (f scriptFormatter) formatDiagnostic(files *token.FileSet, diagnostic analysis.Diagnostic) string {
	source := Span{Start: files.Position(diagnostic.Pos), End: files.Position(diagnostic.End)}
	message := strings.ReplaceAll(diagnostic.Message, f.root+string(filepath.Separator), "")
	return fmt.Sprintf("%s [%s] %s\n", f.span(source), diagnostic.Category, message)
}

func (f scriptFormatter) position(pos token.Position) string {
	if !pos.IsValid() {
		return "-"
	}
	name, err := filepath.Rel(f.root, pos.Filename)
	f.ts.Check(err)
	return fmt.Sprintf("%s:%d:%d", filepath.ToSlash(name), pos.Line, pos.Column)
}

func (f scriptFormatter) span(source Span) string {
	return f.position(source.Start) + ".." + f.position(source.End)
}
