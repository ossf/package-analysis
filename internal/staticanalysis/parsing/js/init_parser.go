package js

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/ossf/package-analysis/internal/utils"
)

// babelParser holds the content of the parser script.
//
//go:embed babel-parser.js
var babelParser []byte

// packageJSON holds the content of the NPM package.json file, with information
// about the dependencies for the parser
//
//go:embed package.json
var packageJSON []byte

// packageLockJSON holds the content of the NPM package-lock.json file, with
// information about versions and hashes of dependencies for the parser
//
//go:embed package-lock.json
var packageLockJSON []byte

const parserFileName = "babel-parser.js"
const packageJSONFileName = "package.json"
const packageLockJSONFileName = "package-lock.json"

type ParserConfig struct {
	InstallDir string
	ParserPath string
}

type parserFile struct {
	name         string
	contents     []byte
	isExecutable bool
}

var parserFiles = []parserFile{
	{parserFileName, babelParser, true},
	{packageJSONFileName, packageJSON, false},
	{packageLockJSONFileName, packageLockJSON, false},
}

func InitParser(installDir string) (ParserConfig, error) {
	if err := os.MkdirAll(installDir, 0o777); err != nil {
		return ParserConfig{}, fmt.Errorf("error creating JS parser directory: %v", err)
	}

	for _, file := range parserFiles {
		filePath := path.Join(installDir, file.name)
		if err := utils.WriteFile(filePath, file.contents, file.isExecutable); err != nil {
			return ParserConfig{}, fmt.Errorf("error writing %s to %s: %v", file.name, installDir, err)
		}
	}

	// run npm install in that folder
	cmd := exec.Command("npm", "ci", "--silent", "--no-progress", "--prefix", installDir, "install")

	if err := cmd.Run(); err != nil {
		return ParserConfig{}, fmt.Errorf("npm install error: %v", err)
	}

	return ParserConfig{
		InstallDir: installDir,
		ParserPath: path.Join(installDir, parserFileName),
	}, nil
}
