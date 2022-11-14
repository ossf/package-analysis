package pkgecosystem

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type npmJSON struct {
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
}

func getNpmLatest(pkg string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", pkg))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details npmJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", err
	}

	return details.DistTags.Latest, nil
}

func makeNpmPackageSpec(pkgName, version string) string {
	if version == "" {
		return pkgName
	} else {
		return pkgName + "@" + version
	}
}

// npmPackData represents the JSON format output by `npm pack`
// when it successfully downloads a package archive.
// The format was found by inspecting JSON output of the NPM CLI.
type npmPackData struct {
	PackageId    string        `json:"id"`
	Name         string        `json:"name"`
	Version      string        `json:"version"`
	Size         int           `json:"size"`
	UnpackedSize int           `json:"unpackedSize"`
	ShaSum       string        `json:"shasum"`
	Integrity    string        `json:"integrity"`
	Filename     string        `json:"filename"`
	Files        []npmFileData `json:"files"`
	EntryCount   int           `json:"entryCount"`
	Bundled      interface{}   `json:"bundled"` // not needed
}

// npmPackData represents the JSON format of a single file
// listed in an NPM package archive, as part of npmPackData.
// The format was found by inspecting JSON output of the NPM CLI.
type npmFileData struct {
	Path string `json:"path"`
	Size int    `json:"size"`
	Mode int    `json:"mode"`
}

// npmPackError represents the JSON format output by `npm pack`
// when an error occurs while attempting to download a package archive.
// The format was found by inspecting JSON output of the NPM CLI.
type npmPackError struct {
	Error struct {
		Code    string `json:"code"`
		Summary string `json:"summary"`
		Detail  string `json:"detail"`
	} `json:"error"`
}

func downloadNpmArchive(pkgName string, version string, directory string) (path string, err error) {
	funcName := "downloadNpmArchive"
	command := "npm"
	if directory == "" {
		// NPM defaults to current directory to download package archives, so we'll do the same
		directory = "."
	}

	npmArgs := []string{
		"pack",
		"--silent",
		"--json",
		"--pack-destination",
		directory,
		makeNpmPackageSpec(pkgName, version),
	}

	cmd := exec.Command(command, npmArgs...)

	var out []byte
	out, err = cmd.Output()

	if err != nil {
		// If the command exited abnormally, it might be an NPM error, in which case we can
		// parse the output JSON to determine the cause of the error. If not, it is probably
		// some kind of OS error in which case we just want to return whatever error we get.
		if exitErr, ok := err.(*exec.ExitError); ok {
			errorOutput := string(exitErr.Stderr)
			// attempt to parse it as JSON
			decoder := json.NewDecoder(strings.NewReader(errorOutput))
			var errorData npmPackError
			decodeErr := decoder.Decode(&errorData)
			if decodeErr == nil {
				errorCode := errorData.Error.Code
				errorDescription := errorData.Error.Summary
				return "", fmt.Errorf("%s: [NPM] %s: %s", funcName, errorCode, errorDescription)
			}
		}
		return "", fmt.Errorf("%s: exec error (%v)", funcName, err)
	}

	if len(out) == 0 {
		return "", fmt.Errorf("%s: no output from NPM process", funcName)
	}

	outputJson := string(out)
	var packData []npmPackData
	decoder := json.NewDecoder(strings.NewReader(outputJson))
	decodeErr := decoder.Decode(&packData)

	if decodeErr != nil {
		return "", fmt.Errorf("%s: could not decode the following JSON output ufrom NPM\n%s", funcName, outputJson)
	}

	if len(packData) == 0 {
		return "", fmt.Errorf("%s: NPM package data is empty", funcName)
	}

	if len(packData) > 1 {
		return "", fmt.Errorf(
			"%s: NPM returned data for %d packages, expected 1\n%v", funcName, len(packData), packData)
	}

	// TODO verify SHA sums

	archiveName := packData[0].Filename
	archivePath := fmt.Sprintf("%s%s%s", directory, string(os.PathSeparator), archiveName)

	return archivePath, nil
}

var npmPkgManager = PkgManager{
	name:    "npm",
	image:   "gcr.io/ossf-malware-analysis/node",
	command: "/usr/local/bin/analyze.js",
	latest:  getNpmLatest,
	runPhases: []RunPhase{
		Install,
		Import,
	},
}
