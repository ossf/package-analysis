package analysis

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type npmJSON struct {
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
}

func getNPMLatest(pkg string) string {
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", pkg))
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details npmJSON
	err = decoder.Decode(&details)
	if err != nil {
		log.Panic(err)
	}

	return details.DistTags.Latest
}

var NPMPackageManager = PkgManager{
	Image: "gcr.io/ossf-malware-analysis/node",
	CommandFmt: func(pkg, ver string) string {
		if ver != "" {
			return fmt.Sprintf("analyze.js %s@%s", pkg, ver)
		}

		return fmt.Sprintf("analyze.js %s", pkg)
	},
	GetLatest: getNPMLatest,
}
