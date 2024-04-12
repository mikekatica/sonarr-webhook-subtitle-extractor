package main

import (
	"flag"
	"github.com/mikekatica/sonarr-webhook-subtitle-extractor/services/webservices"

	"github.com/golang/glog"
)

func main() {
	var bindaddr = flag.String("bindaddr", ":8111", "Location of the file to extract subtitles from")
	var connection = flag.String("pguri", "REQUIRED", "pgsql connection string")
	flag.Parse()
	glog.Infof("Running mkv subtitle extractor webservice on %v...", *bindaddr)
	glog.Infof("Using postgres connection %v...", *connection)
	g := webservices.New(*bindaddr, *connection)
	g.Serve()
}
