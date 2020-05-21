package main

import (
	"flag"
	"net/url"
)

var config struct {
	gitHubRepo string
	startTag   string
	endTag     string
}

func gitHubOrgProject() string {
	u, _ := url.Parse(config.gitHubRepo)
	return u.Path[:len(u.Path)-4]
}

func init() {
	flag.StringVar(&config.gitHubRepo, "repo", "https://github.com/Kong/kuma.git", "The GitHub repo to process")
	flag.StringVar(&config.startTag, "start", "", "The start hash or tag")
	flag.StringVar(&config.endTag, "end", "", "The end hash or tag")
}
