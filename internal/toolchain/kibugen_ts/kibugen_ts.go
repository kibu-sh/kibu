package kibugen_ts

import (
	"strings"

	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name:     "kibugen_ts",
	Doc:      "Analyzes go source code for kibu services and generates typescript",
	Requires: []*analysis.Analyzer{kibumod.Analyzer},
	//ResultType:       resultType,
	RunDespiteErrors: true,
	Run:              run,
}

var missingPackageError = errors.New("missing result of kibugen_ts analyzer")

func run(pass *analysis.Pass) (any, error) {
	pkg, ok := kibumod.FromPass(pass)
	if !ok {
		return nil, missingPackageError
	}

	//if len(pkg.Services) == 0 {
	//	return nil, nil
	//}

	s := GenerateTypeScript(pkg)
	_ = s

	return nil, nil
}

func GenerateTypeScript(pkg *modspecv2.Package) string {
	var sb strings.Builder

	for _, svc := range pkg.Services {
		sb.WriteString("export type ")
		sb.WriteString(svc.Name)
		sb.WriteString(" = {\n")

		for _, op := range svc.Operations {
			sb.WriteString("  ")
			sb.WriteString(op.Name)
			sb.WriteString("() {}\n")
		}

		sb.WriteString("}\n")
	}

	return sb.String()
}

func Main() (int, error) {
	return 0, nil
}
