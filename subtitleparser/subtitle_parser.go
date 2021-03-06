package subtitleparser

import (
	"os"
	"time"

	"github.com/golang/glog"

	"github.com/remko/go-mkvparse"
)

const (
	VIDEO    int64 = 1
	AUDIO    int64 = 2
	COMPLEX  int64 = 3
	LOGO     int64 = 16
	SUBTITLE int64 = 17
	BUTTONS  int64 = 18
	CONTROL  int64 = 32
)

const (
	SSA string = "S_TEXT/SSA"
	ASS string = "S_TEXT/ASS"
)

var defaultlang = "eng"

type SubtitleTrack struct {
	TrackID  int64
	Language string
	Codec    string
	Default  bool
	Forced   bool
}

type SubtitleTrackHandler struct {
	mkvparse.Handler
	currentint64     int64
	currentTrackID   int64
	currentLanguage  *string
	currentTrackUID  int64
	currentCodec     *string
	currentIsDefault bool
	currentForced    bool
	Subtitles        map[int64]SubtitleTrack
}

func (p *SubtitleTrackHandler) HandleMasterBegin(id mkvparse.ElementID, info mkvparse.ElementInfo) (bool, error) {
	glog.Infof("MasterBegin: Got element ID of %v", id)
	if id == mkvparse.TrackEntryElement {
		glog.V(4).Infof("Found a subtitle, Descending into element")
		p.currentLanguage = &defaultlang
		p.currentCodec = nil
	}
	return true, nil
}
func (p *SubtitleTrackHandler) HandleMasterEnd(id mkvparse.ElementID, info mkvparse.ElementInfo) error {
	glog.V(4).Infof("MasterEnd: Got element ID of %v", id)
	if id == mkvparse.TrackEntryElement && p.currentint64 == SUBTITLE {
		sub := SubtitleTrack{
			TrackID:  p.currentTrackID,
			Language: *p.currentLanguage,
			Codec:    *p.currentCodec,
			Default:  p.currentIsDefault,
			Forced:   p.currentForced,
		}
		p.Subtitles[p.currentTrackUID] = sub
	}
	return nil
}

func (p *SubtitleTrackHandler) HandleString(id mkvparse.ElementID, value string, info mkvparse.ElementInfo) error {
	switch id {
	case mkvparse.LanguageElement:
		glog.V(4).Infof("Found a language for the track of %v", value)
		p.currentLanguage = &value
	case mkvparse.LanguageIETFElement:
		glog.V(4).Infof("Found an ietf language for the track of %v", value)
	case mkvparse.CodecIDElement:
		glog.V(4).Infof("Found a codec id for the track of %v", value)
		p.currentCodec = &value
	}
	return nil
}

func (p *SubtitleTrackHandler) HandleInteger(id mkvparse.ElementID, value int64, info mkvparse.ElementInfo) error {
	switch id {
	case mkvparse.TrackUIDElement:
		p.currentTrackUID = value
	case mkvparse.TrackTypeElement:
		p.currentint64 = value
	case mkvparse.TrackNumberElement:
		p.currentTrackID = value
	case mkvparse.FlagDefaultElement:
		p.currentIsDefault = value != 0
	case mkvparse.FlagForcedElement:
		p.currentForced = value != 0
	}
	return nil
}

func (p *SubtitleTrackHandler) HandleFloat(id mkvparse.ElementID, value float64, info mkvparse.ElementInfo) error {
	return nil
}

func (p *SubtitleTrackHandler) HandleDate(id mkvparse.ElementID, value time.Time, info mkvparse.ElementInfo) error {
	return nil
}

func (p *SubtitleTrackHandler) HandleBinary(id mkvparse.ElementID, value []byte, info mkvparse.ElementInfo) error {
	return nil
}

func ExtractSubtitleInfo(filepath string) (map[int64]SubtitleTrack, error) {
	glog.Infof("Extracting subtitles from %v", filepath)
	file, _ := os.Open(filepath)
	defer file.Close()
	h := SubtitleTrackHandler{
		Subtitles: make(map[int64]SubtitleTrack),
	}
	glog.V(4).Info("Parsing sub tracks from mkv")
	err := mkvparse.ParseSections(file, &h, mkvparse.TracksElement)
	if err != nil {
		glog.Error(err)
	}
	return h.Subtitles, nil
}
