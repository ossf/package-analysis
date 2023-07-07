package dynamicanalysis

import (
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

// defaultCommand returns the path (in the default sandbox image)
// of the default dynamic analysis command for the ecosystem
var defaultCommand = map[pkgecosystem.Ecosystem]string{
	pkgecosystem.CratesIO:  "/usr/local/bin/analyze-rust.py",
	pkgecosystem.NPM:       "/usr/local/bin/analyze-node.js",
	pkgecosystem.Packagist: "/usr/local/bin/analyze-php.php",
	pkgecosystem.PyPI:      "/usr/local/bin/analyze-python.py",
	pkgecosystem.RubyGems:  "/usr/local/bin/analyze-ruby.rb",
}

func DefaultCommand(ecosystem pkgecosystem.Ecosystem) string {
	cmd := defaultCommand[ecosystem]
	if cmd == "" {
		panic("unsupported ecosystem: " + ecosystem)
	}
	return cmd
}

// MakeAnalysisArgs returns the arguments to pass to the dynamic analysis command in the sandbox
// for the given phase of dynamic analysis on a package. The actual analysis command
// depends on the ecosystem, see pkgmanager.PkgManager.DynamicAnalysisCommand()
func MakeAnalysisArgs(p *pkgmanager.Pkg, phase analysisrun.DynamicPhase) []string {
	args := make([]string, 0)

	if p.IsLocal() {
		args = append(args, "--local", p.LocalPath())
	} else if p.Version() != "" {
		args = append(args, "--version", p.Version())
	}

	if phase == "" {
		args = append(args, "all")
	} else {
		args = append(args, string(phase))
	}

	args = append(args, p.Name())

	return args
}
