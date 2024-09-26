package kibugenv2

import (
	"golang.org/x/tools/go/analysis"
	"os"
	"path/filepath"
	"strings"
)

func SaveArtifacts(rootDir string, results []*analysis.Pass) ([]string, error) {
	var outFiles []string
	for _, pass := range results {
		artifact, ok := FromPass(pass)
		if !ok {
			continue
		}

		outfile, err := saveArtifact(rootDir, artifact, pass)
		if err != nil {
			return outFiles, err
		}

		outFiles = append(outFiles, outfile)
	}
	return outFiles, nil
}

func saveArtifact(dir string, artifact *Artifact, pass *analysis.Pass) (string, error) {
	pkgDir := PackagePathFromAnalysis(dir, pass)
	filename := filepath.Join(pkgDir, fileWithGenGoExt(pass.Pkg.Name()))
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	return filename, artifact.File.Render(file)
}

func PackagePathFromAnalysis(rootDir string, pass *analysis.Pass) string {
	return filepath.Join(rootDir, strings.Replace(pass.Pkg.Path(), pass.Module.Path, "", 1))
}
