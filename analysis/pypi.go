package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type pypiJSON struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
}

func getPyPILatest(pkg string) string {
	resp, err := http.Get(fmt.Sprintf("https://pypi.org/pypi/%s/json", pkg))
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details pypiJSON
	err = decoder.Decode(&details)
	if err != nil {
		log.Panic(err)
	}

	return details.Info.Version
}

var PyPIPackageManager = pkgManager{
	image: "gcr.io/ossf-malware-analysis/python",
	commandFmt: func(pkg, ver string) string {
		return fmt.Sprintf("analyze.py %s==%s", pkg, ver)
	},
	getLatest: getPyPILatest,
}
