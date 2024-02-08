package staticanalysis

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/staticanalysis/basicdata"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/signals"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
	"github.com/ossf/package-analysis/pkg/valuecounts"
)

type testFile struct {
	filename    string
	contents    []byte
	sha256      string
	fileType    string
	lineLengths valuecounts.ValueCounts
}

var helloWorldJs = testFile{
	filename:    "hi.js",
	contents:    []byte(`console.log("hi");` + "\n"),
	sha256:      "2bf8b125d15a71b5fa79fe710cae0db911a71e65891e270bca1d4eb5dd785288",
	fileType:    "ASCII text",
	lineLengths: valuecounts.Count([]int{18}),
}

func makeDesiredResult(files ...testFile) []SingleResult {
	result := make([]SingleResult, len(files))
	for index, file := range files {
		result[index] = SingleResult{
			Filename: file.filename,
			Basic: &basicdata.FileData{
				DetectedType: file.fileType,
				Size:         int64(len(file.contents)),
				SHA256:       file.sha256,
				LineLengths:  file.lineLengths,
			},
			Parsing: &parsing.SingleResult{
				Language:    parsing.JavaScript,
				Identifiers: []token.Identifier{},
				StringLiterals: []token.String{
					{Value: "hi", Raw: `"hi"`, Entropy: math.Log(2.0)},
				},
				IntLiterals:   []token.Int{},
				FloatLiterals: []token.Float{},
				Comments:      []token.Comment{},
			},
			Signals: &signals.FileSignals{
				IdentifierLengths:     valuecounts.New(),
				StringLengths:         valuecounts.Count([]int{2}),
				SuspiciousIdentifiers: []staticanalysis.SuspiciousIdentifier{},
				EscapedStrings:        []staticanalysis.EscapedString{},
				Base64Strings:         []string{},
				HexStrings:            []string{},
				IPAddresses:           []string{},
				URLs:                  []string{},
			},
		}
	}

	return result
}

func TestAnalyzePackageFiles(t *testing.T) {
	tests := []struct {
		name    string
		files   []testFile
		want    []SingleResult
		wantErr bool
	}{
		{
			name:    "hello JS",
			files:   []testFile{helloWorldJs},
			want:    makeDesiredResult(helloWorldJs),
			wantErr: false,
		},
	}
	parserDir := t.TempDir()
	jsParserConfig, err := parsing.InitParser(context.Background(), parserDir)
	if err != nil {
		t.Errorf("failed to init parser: %v", err)
		return
	}

	log.Initialize("")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractDir := t.TempDir()
			for _, file := range tt.files {
				extractPath := filepath.Join(extractDir, file.filename)
				fmt.Printf("writing %s to %s\n", file.filename, extractPath)
				if err := os.WriteFile(extractPath, file.contents, 0o666); err != nil {
					t.Errorf("failed to write test file: %v", err)
					return
				}
			}
			got, err := AnalyzePackageFiles(context.Background(), extractDir, jsParserConfig, AllTasks())
			if (err != nil) != tt.wantErr {
				t.Errorf("AnalyzePackageFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AnalyzePackageFiles() \n"+
					"------------- got ----------------\n%v\n\n"+
					"--------------want----------------\n%v", got, tt.want)
			}
		})
	}
}
