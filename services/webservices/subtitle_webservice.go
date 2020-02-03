package webservice

import (
	"net/http"
	"os"
	"path"
	"sonarr-webhook-subtitle-extractor/services/types"
	"sonarr-webhook-subtitle-extractor/subtitleparser"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/glog"
)

type SubtitleWebservice struct {
	Engine      *gin.Engine
	BindAddress string
}

type SimpleExtractRequest struct {
	Filepath string `json:"filepath"`
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

func (w *SubtitleWebservice) ExtractSubtitleSonarrAPI() gin.HandlerFunc {
	return func(context *gin.Context) {
		var event types.SonarrEvent
		err := context.Copy().MustBindWith(&event, binding.JSON)
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
		err := context.Copy().MustBindWith(&event, binding.JSON)
		glog.V(4).Infof("Recieved a request: %v", event)
		if err != nil {
			glog.Errorf("Received an error processing the request: %v", err)
			return
		}
		context.JSON(http.StatusOK, gin.H{})
		go w.ExtractSubtitleAction(event.Filepath)
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
	return &svc
}

func (w *SubtitleWebservice) Serve() error {
	return w.Engine.Run(w.BindAddress)
}
