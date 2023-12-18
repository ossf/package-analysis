package parsing

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

const (
	parserFileName          = "babel-parser.js"
	packageJSONFileName     = "package.json"
	packageLockJSONFileName = "package-lock.json"
)

// npmCacheDir is used to check for cached versions of NPM dependencies before
// downloading them from a remote source. The directory is populated by the
// Docker build for the container this code will run in.
const npmCacheDir = "/npm_cache"

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
	{parserFileName, babelParser, false},
	{packageJSONFileName, packageJSON, false},
	{packageLockJSONFileName, packageLockJSON, false},
}

func InitParser(ctx context.Context, installDir string) (ParserConfig, error) {
	if err := os.MkdirAll(installDir, 0o777); err != nil {
		return ParserConfig{}, fmt.Errorf("error creating JS parser directory: %w", err)
	}

	for _, file := range parserFiles {
		writePath := filepath.Join(installDir, file.name)
		if err := utils.WriteFile(writePath, file.contents, file.isExecutable); err != nil {
			return ParserConfig{}, fmt.Errorf("error writing %s to %s: %w", file.name, installDir, err)
		}
	}

	// run npm install in that folder
	npmArgs := []string{"ci", "--silent", "--no-progress", "--prefix", installDir}

	fileInfo, err := os.Stat(npmCacheDir)
	cacheDirAccessible := err == nil && fileInfo.IsDir() && (fileInfo.Mode().Perm()&0o700 == 0o700)
	if cacheDirAccessible {
		npmArgs = append(npmArgs, "--cache", npmCacheDir, "--prefer-offline")
	}

	cmd := exec.CommandContext(ctx, "npm", npmArgs...)
	if err := cmd.Run(); err != nil {
		return ParserConfig{}, fmt.Errorf("npm install error: %w", err)
	}

	return ParserConfig{
		InstallDir: installDir,
		ParserPath: filepath.Join(installDir, parserFileName),
	}, nil
}
