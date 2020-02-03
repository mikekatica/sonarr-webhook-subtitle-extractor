package webservice

import (
	"net/http"
	"sonarr-webhook-subtitle-extractor/subtitleparser"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/glog"
)

type SubtitleWebservice struct {
	Engine      *gin.Engine
	BindAddress string
}

type ExtractSubtitlesRequest struct {
	Filepath string `json:"filepath"`
}

func (w *SubtitleWebservice) ExtractSubtitleAction(filepath string) {
	subs, err := subtitleparser.ExtractSubtitleInfo(filepath)
	if err != nil {
		glog.Errorf("Couldn't extract sub track info from %v: %v", filepath, err)
	}
	trackUID, err := subtitleparser.DecideSubtitleTrack(subs)
	if err != nil {
		glog.Errorf("Couldn't decide subs from %v: %v", filepath, err)
	}
	err = subtitleparser.ExtractSubtitleFromFile(filepath, subs[*trackUID])
	if err != nil {
		glog.Errorf("Couldn't extract subs from %v: %v", filepath, err)
	}
}

func (w *SubtitleWebservice) ExtractSubtitleAPI() gin.HandlerFunc {
	return func(context *gin.Context) {
		var reqBody ExtractSubtitlesRequest
		err := context.MustBindWith(&reqBody, binding.JSON)
		if err == nil {
			glog.V(4).Infof("Recieved a request: %v", reqBody)
			context.JSON(http.StatusOK, gin.H{})
			context.Next()
			return
		}
	}
}

func New(bindaddr string) *SubtitleWebservice {
	r := gin.Default()
	svc := SubtitleWebservice{
		Engine:      r,
		BindAddress: bindaddr,
	}
	r.POST("/extract", svc.ExtractSubtitleAPI())
	return &svc
}

func (w *SubtitleWebservice) Serve() error {
	return w.Engine.Run(w.BindAddress)
}
