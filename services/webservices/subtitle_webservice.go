package webservices

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"github.com/mikekatica/sonarr-webhook-subtitle-extractor/services/types"
	"github.com/mikekatica/sonarr-webhook-subtitle-extractor/subtitleparser"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/glob"
	"github.com/golang/glog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type SubtitleWebservice struct {
	Engine      *gin.Engine
	BindAddress string
	DbEngine    *bun.DB
}

type SimpleExtractRequest struct {
	Filepath string `json:"filepath"`
}

type BulkExtractRequest struct {
	Basepath      string `json:"basepath"`
	FileGlob      string `json:"fileglob"`
	TrackOverride int64  `json:"tracknum"`
}

type SubtitleExtractResult struct{
	bun.BaseModel `bun:"table:subtitle_extract_results,alias:ser"`

	ID      int64  `bun:"id,pk,autoincrement"`
	File    string `bun:"file,notnull"`
	Result  bool   `bun:"result,notnull"`
	Message string `bun:"message,notnull"`
}

func (w *SubtitleWebservice) PingDb() error {
	return w.DbEngine.Ping()
}

func (w *SubtitleWebservice) ExtractSubtitleAction(lang subtitleparser.SubtitleLanguageDefault, filepath string) {
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
	var res SubtitleExtractResult
	res.File = filepath
	deadline := time.Now().Add(60 * time.Second)
	defer w.DbEngine.NewInsert().Model(&res).Exec(context.WithDeadline(context.Background(), deadline))
	select {
	case _ = <-waitForFile:
		glog.Infof("File %v exists and is written", filepath)
	case <-time.After(15 * time.Minute):
		glog.Warningf("File %v does not exist after 15 minute timeout", filepath)
		res.Result = false
		res.Message = "File does not exist after 15 minute timeout"
		return
	}
	subs, err := subtitleparser.ExtractSubtitleInfo(filepath)
	if err != nil {
		glog.Errorf("Couldn't extract sub track info from %v: %v", filepath, err)
		res.Result = false
		res.Message = fmt.Sprintf("Couldn't extract sub track info from %v: %v", filepath, err)
		return
	}
	track, err := subtitleparser.DecideSubtitleTrack(lang, subs)
	if err != nil {
		glog.Errorf("Couldn't decide subs from %v: %v", filepath, err)
		res.Result = false
		res.Message = fmt.Sprintf("Couldn't decide subs from %v: %v", filepath, err)
		return
	}
	err = subtitleparser.ExtractSubtitleFromFile(filepath, track)
	if err != nil {
		glog.Errorf("Couldn't extract subs from %v: %v", filepath, err)
		res.Result = false
		res.Message = fmt.Sprintf("Couldn't extract subs from %v: %v", filepath, err)
		return
	}
	res.Result = true
	res.Message = "Extraction Succeeded"
	return
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
		defaultLang := strings.ReplaceAll(context.Param("lang"), "/", "")
		var lang *string
		lang = nil
		if defaultLang != "" {
			glog.V(4).Infof("Looking for language: %v", defaultLang)
			lang = &defaultLang
		}
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
			go w.ExtractSubtitleAction(subtitleparser.SubtitleLanguageDefault{DefaultLang: lang}, copiedFilePath)
		} else if event.EventType == "Test" {
			glog.Infof("Sonarr is testing, and it is working: %v", event)
		}
	}
}

func (w *SubtitleWebservice) ExtractSubtitleSimpleAPI() gin.HandlerFunc {
	return func(context *gin.Context) {
		defaultLang := strings.ReplaceAll(context.Param("lang"), "/", "")
		var lang *string
		lang = nil
		if defaultLang != "" {
			glog.V(4).Infof("Looking for language: %v", defaultLang)
			lang = &defaultLang
		}
		var event SimpleExtractRequest
		err := context.Copy().ShouldBindJSON(&event)
		glog.V(4).Infof("Recieved a request: %v", event)
		if err != nil {
			glog.Errorf("Received an error processing the request: %v", err)
			return
		}
		context.JSON(http.StatusOK, gin.H{})
		go w.ExtractSubtitleAction(subtitleparser.SubtitleLanguageDefault{DefaultLang: lang}, event.Filepath)
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

func New(bindaddr string, connectionstring string) *SubtitleWebservice {
	r := gin.Default()
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connectionstring)))
	engine := bun.NewDB(sqldb, pgdialect.New())
	svc := SubtitleWebservice{
		Engine:      r,
		BindAddress: bindaddr,
		DbEngine:    engine,
	}
	if err := svc.PingDb(); err != nil {
		glog.Errorf("Could not connect to database: %v", err)
		panic(err)
	}
	deadline := time.Now().Add(60 * time.Second)
	svc.DbEngine.NewCreateTable().Model((*SubtitleExtractResult)(nil)).Exec(context.WithDeadline(context.Background(), deadline))
	r.LoadHTMLGlob("public/*")
	r.POST("/extract/sonarr/*lang", svc.ExtractSubtitleSonarrAPI())
	r.POST("/extract/simple/*lang", svc.ExtractSubtitleSimpleAPI())
	r.POST("/extract/bulk", svc.ExtractSubtitleBulkAPI())
	r.GET("/results", func (c *gin.Context) {
		cdeadline := time.Now().Add(60 * time.Second)
		var res SubtitleExtractResult
		err := svc.DbEngine.NewSelect().Model(res).Order("id DESC").Limit(20).Scan(context.WithDeadline(context.Background(), cdeadline))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{"results": res})
	})
	return &svc
}

func (w *SubtitleWebservice) Serve() error {
	return w.Engine.Run(w.BindAddress)
}
