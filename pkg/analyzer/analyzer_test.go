package analyzer_test

import (
	"testing"

	"github.com/piotrpersona/slen/pkg/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	pkgs := []string{
		"main",
	}
	analysistest.Run(t, analysistest.TestData(), analyzer.Analyzer, pkgs...)
}
