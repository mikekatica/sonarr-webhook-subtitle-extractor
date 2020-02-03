package webservice

import (
	"net/http"
	"sonarr-webhook-subtitle-extractor/services/types"
	"sonarr-webhook-subtitle-extractor/subtitleparser"

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
	subs, err := subtitleparser.ExtractSubtitleInfo(filepath)
	if err != nil {
		glog.Errorf("Couldn't extract sub track info from %v: %v", filepath, err)
	}
	track, err := subtitleparser.DecideSubtitleTrack(subs)
	if err != nil {
		glog.Errorf("Couldn't decide subs from %v: %v", filepath, err)
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
			go w.ExtractSubtitleAction(event.EpisodeFile.Path)
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
