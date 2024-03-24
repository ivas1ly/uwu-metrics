package exitcall_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/ivas1ly/uwu-metrics/pkg/staticlint/exitcall"
)

func Test(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), exitcall.Analyzer, "./...")
}
