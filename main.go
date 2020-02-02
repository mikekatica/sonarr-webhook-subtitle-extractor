package main

import (
	"flag"
	"sonarr-webhook-subtitle-extractor/services/webservices"

	"github.com/golang/glog"
)

func main() {
	var bindaddr = flag.String("bindaddr", ":8111", "Location of the file to extract subtitles from")
	flag.Parse()
	glog.Infof("Running mkv subtitle extractor webservice on %v...", *bindaddr)
	g := webservice.New(*bindaddr)
	g.Serve()
}
