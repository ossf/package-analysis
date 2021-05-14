package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type rubygemsJSON struct {
	Version string `json:"version"`
}

func getRubyGemsLatest(pkg string) string {
	resp, err := http.Get(fmt.Sprintf("https://rubygems.org/api/v1/gems/%s.json", pkg))
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details rubygemsJSON
	err = decoder.Decode(&details)
	if err != nil {
		log.Panic(err)
	}

	return details.Version
}

var RubyGemsPackageManager = pkgManager{
	image: "gcr.io/ossf-malware-analysis/ruby",
	commandFmt: func(pkg, ver string) string {
		return fmt.Sprintf("analyze.rb %s %s", pkg, ver)
	},
	getLatest: getRubyGemsLatest,
}
