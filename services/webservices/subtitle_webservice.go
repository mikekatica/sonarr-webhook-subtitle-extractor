package webservices

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sonarr-webhook-subtitle-extractor/services/types"
	"sonarr-webhook-subtitle-extractor/subtitleparser"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/glob"
	"github.com/golang/glog"
)

type SubtitleWebservice struct {
	Engine      *gin.Engine
	BindAddress string
}

type SimpleExtractRequest struct {
	Filepath string `json:"filepath"`
}

type BulkExtractRequest struct {
	Basepath      string `json:"basepath"`
	FileGlob      string `json:"fileglob"`
	TrackOverride int64  `json:"tracknum"`
}

func (w *SubtitleWebservice) ExtractSubtitleAction(filepath string) {
	waitForFile := make(chan os.FileInfo, 1)
	var fileStatErr error
	glog.Infof("Waiting for %v to exist.", filepath)
	go func() {
		for _, fileStatErr = os.Stat(filepath); os.IsNotExist(fileStatErr); _, fileStatErr = os.Stat(filepath) {
			time.Sleep(300 * time.Millisecond)
		}
		finfo, _ := os.Stat(filepath)
		waitForFile <- finfo
	}()
	select {
	case _ = <-waitForFile:
		glog.Infof("File %v exists and is written", filepath)
	case <-time.After(15 * time.Minute):
		glog.Warningf("File %v does not exist after 15 minute timeout", filepath)
		return
	}
	subs, err := subtitleparser.ExtractSubtitleInfo(filepath)
	if err != nil {
		glog.Errorf("Couldn't extract sub track info from %v: %v", filepath, err)
	}
	track, err := subtitleparser.DecideSubtitleTrack(subs)
	if err != nil {
		glog.Errorf("Couldn't decide subs from %v: %v", filepath, err)
		return
	}
	err = subtitleparser.ExtractSubtitleFromFile(filepath, track)
	if err != nil {
		glog.Errorf("Couldn't extract subs from %v: %v", filepath, err)
	}
}

func (w *SubtitleWebservice) ExtractSubtitleAction2(filepath string, track int64) {
	waitForFile := make(chan os.FileInfo, 1)
	var fileStatErr error
	glog.Infof("Waiting for %v to exist.", filepath)
	go func() {
		for _, fileStatErr = os.Stat(filepath); os.IsNotExist(fileStatErr); _, fileStatErr = os.Stat(filepath) {
			time.Sleep(300 * time.Millisecond)
		}
		finfo, _ := os.Stat(filepath)
		waitForFile <- finfo
	}()
	select {
	case _ = <-waitForFile:
		glog.Infof("File %v exists and is written", filepath)
	case <-time.After(15 * time.Minute):
		glog.Warningf("File %v does not exist after 15 minute timeout", filepath)
		return
	}
	subs, err := subtitleparser.ExtractSubtitleInfo(filepath)
	if err != nil {
		glog.Errorf("Couldn't extract sub track info from %v: %v", filepath, err)
	}
	glog.V(4).Infof("Found subtitles: %v", subs)
	sub, err := subtitleparser.GetSubtitleTrack(subs, track)
	if err != nil {
		glog.Errorf("Couldn't decide subs from %v: %v", filepath, err)
		return
	}
	err = subtitleparser.ExtractSubtitleFromFile(filepath, sub)
	if err != nil {
		glog.Errorf("Couldn't extract subs from %v: %v", filepath, err)
	}
}

func (w *SubtitleWebservice) ExtractSubtitleSonarrAPI() gin.HandlerFunc {
	return func(context *gin.Context) {
		var event types.SonarrEvent
		err := context.Copy().ShouldBindJSON(&event)
		glog.V(4).Infof("Recieved a request: %v", event)
		if err != nil {
			glog.Errorf("Received an error processing the request: %v", err)
			return
		}
		context.JSON(http.StatusOK, gin.H{})
		if event.EventType != nil && event.EventType == "Download" {
			glog.V(5).Infof("Series Info: %v", event.Series)
			glog.V(5).Infof("Episode File Info: %v", event.EpisodeFile)
			copiedFilePath := path.Join(event.Series.Path, event.EpisodeFile.RelativePath)
			go w.ExtractSubtitleAction(copiedFilePath)
		} else if event.EventType == "Test" {
			glog.Infof("Sonarr is testing, and it is working: %v", event)
		}
	}
}

func (w *SubtitleWebservice) ExtractSubtitleSimpleAPI() gin.HandlerFunc {
	return func(context *gin.Context) {
		var event SimpleExtractRequest
		err := context.Copy().ShouldBindJSON(&event)
		glog.V(4).Infof("Recieved a request: %v", event)
		if err != nil {
			glog.Errorf("Received an error processing the request: %v", err)
			return
		}
		context.JSON(http.StatusOK, gin.H{})
		go w.ExtractSubtitleAction(event.Filepath)
	}
}

func (w *SubtitleWebservice) ExtractSubtitleBulkAPI() gin.HandlerFunc {
	return func(context *gin.Context) {
		var event BulkExtractRequest
		err := context.Copy().ShouldBindJSON(&event)
		glog.V(4).Infof("Recieved a request: %v", event)
		if err != nil {
			glog.Errorf("Received an error processing the request: %v", err)
			return
		}
		context.JSON(http.StatusOK, gin.H{})
		fileglob := glob.MustCompile(event.FileGlob)
		filepath.Walk(event.Basepath,
			func(path string, info os.FileInfo, err error) error {
				if fileglob.Match(path) {
					if err != nil {
						glog.Errorf("Found a file to process at %v but there was an error: %v", path, err)
						return err
					}
					glog.V(4).Infof("Processing file found at: %v", path)
					go w.ExtractSubtitleAction2(path, event.TrackOverride)
				} else if err != nil {
					return err
				}
				return nil
			})
	}
}

func New(bindaddr string) *SubtitleWebservice {
	r := gin.Default()
	svc := SubtitleWebservice{
		Engine:      r,
		BindAddress: bindaddr,
	}
	r.POST("/extract/sonarr", svc.ExtractSubtitleSonarrAPI())
	r.POST("/extract/simple", svc.ExtractSubtitleSimpleAPI())
	r.POST("/extract/bulk", svc.ExtractSubtitleBulkAPI())
	return &svc
}

func (w *SubtitleWebservice) Serve() error {
	return w.Engine.Run(w.BindAddress)
}
