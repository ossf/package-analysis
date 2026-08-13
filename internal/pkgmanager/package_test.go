package pkgmanager

import (
	"testing"

	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

func TestName(t *testing.T) {
	expectedName := "a-package"
	p := Pkg{name: expectedName}

	actualName := p.Name()
	if actualName != expectedName {
		t.Errorf("Name() = %v; want %v", actualName, expectedName)
	}
}

func TestVersion(t *testing.T) {
	expectedVersion := "1.0.0"
	p := Pkg{version: expectedVersion}

	actualVersion := p.Version()
	if actualVersion != expectedVersion {
		t.Errorf("Version() = %v; want %v", actualVersion, expectedVersion)
	}
}

func TestEcosystem(t *testing.T) {
	expectedEcosystem := pkgecosystem.NPM
	p := Pkg{manager: &PkgManager{ecosystem: expectedEcosystem}}

	actualEcosystem := p.Ecosystem()
	if actualEcosystem != expectedEcosystem {
		t.Errorf("Ecosystem() = %v; want %v", actualEcosystem, expectedEcosystem)
	}
}

func TestEcosystemName(t *testing.T) {
	expectedEcosystemName := "npm"
	p := Pkg{manager: &PkgManager{ecosystem: pkgecosystem.NPM}}

	actualEcosystemName := p.EcosystemName()
	if actualEcosystemName != expectedEcosystemName {
		t.Errorf("EcosystemName() = %v; want %v", actualEcosystemName, expectedEcosystemName)
	}
}

func TestManager(t *testing.T) {
	expectedPkgManager := &PkgManager{ecosystem: pkgecosystem.RubyGems}
	p := Pkg{manager: expectedPkgManager}

	actualPkgManager := p.Manager()
	if actualPkgManager != expectedPkgManager {
		t.Errorf("Manager() = %v; want %v", actualPkgManager, expectedPkgManager)
	}
}

func TestIsLocal(t *testing.T) {
	expectedIsLocal := true
	p := Pkg{local: "some/local/path"}

	actualIsLocal := p.IsLocal()
	if actualIsLocal != expectedIsLocal {
		t.Errorf("IsLocal() = %v; want %v", actualIsLocal, expectedIsLocal)
	}
}

func TestLocalPath(t *testing.T) {
	expectedLocalPath := "some/local/path"
	p := Pkg{local: expectedLocalPath}

	actualLocalPath := p.LocalPath()
	if actualLocalPath != expectedLocalPath {
		t.Errorf("LocalPath() = %v; want %v", actualLocalPath, expectedLocalPath)
	}
}
