package webservice

import (
	"github.com/gin-gonic/gin"
)

type SubtitleWebservice struct {
	Engine      *gin.Engine
	BindAddress string
}

func ExtractSubtitle(context *gin.Context) {

	context.JSON(200, gin.H{})
}

func New(bindaddr string) *SubtitleWebservice {
	r := gin.Default()
	r.POST("/extract", ExtractSubtitle)
	svc := SubtitleWebservice{
		Engine:      r,
		BindAddress: bindaddr,
	}
	return &svc
}

func (w *SubtitleWebservice) Serve() error {
	return w.Engine.Run(w.BindAddress)
}
