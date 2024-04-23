package subtitleparser

import (
	"os"
	"strings"
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

var defaultlang = "en"

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

func indent(n int) string {
	return strings.Repeat("  ", n)
}

func (p *SubtitleTrackHandler) HandleMasterBegin(id mkvparse.ElementID, info mkvparse.ElementInfo) (bool, error) {
	switch id {
	case mkvparse.TrackEntryElement:
		glog.V(4).Infof("Found a Track Entry, Descending into element")
		p.currentLanguage = &defaultlang
		p.currentCodec = nil
		return true, nil
	default:
		glog.V(10).Infof("%s- %s:\n", indent(info.Level), mkvparse.NameForElementID(id))
		return true, nil
	}
}
func (p *SubtitleTrackHandler) HandleMasterEnd(id mkvparse.ElementID, info mkvparse.ElementInfo) error {
	if id == mkvparse.TrackEntryElement && p.currentint64 == SUBTITLE {
		sub := SubtitleTrack{
			TrackID:  p.currentTrackID,
			Language: *p.currentLanguage,
			Codec:    *p.currentCodec,
			Default:  p.currentIsDefault,
			Forced:   p.currentForced,
		}
		glog.V(4).Infof("Processed a subtitle: %+v", sub)
		p.Subtitles[p.currentTrackUID] = sub
	}
	return nil
}

func (p *SubtitleTrackHandler) HandleString(id mkvparse.ElementID, value string, info mkvparse.ElementInfo) error {
	glog.V(10).Infof("%s- %v: %q\n", indent(info.Level), mkvparse.NameForElementID(id), value)
	switch id {
	case mkvparse.LanguageElement:
		glog.V(4).Infof("Found a language for the track of %v", value)
		if p.currentLanguage == nil {
			p.currentLanguage = &value
		}
	case mkvparse.LanguageIETFElement:
		glog.V(4).Infof("Found an ietf language for the track of %v", value)
		p.currentLanguage = &value
	case mkvparse.CodecIDElement:
		glog.V(4).Infof("Found a codec id for the track of %v", value)
		p.currentCodec = &value
	}
	return nil
}

func (p *SubtitleTrackHandler) HandleInteger(id mkvparse.ElementID, value int64, info mkvparse.ElementInfo) error {
	glog.V(10).Infof("%s- %v: %q\n", indent(info.Level), mkvparse.NameForElementID(id), value)
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
	glog.V(10).Infof("%s- %v: %v\n", indent(info.Level), mkvparse.NameForElementID(id), value)
	return nil
}

func (p *SubtitleTrackHandler) HandleDate(id mkvparse.ElementID, value time.Time, info mkvparse.ElementInfo) error {
	glog.V(10).Infof("%s- %v: %v\n", indent(info.Level), mkvparse.NameForElementID(id), value)
	return nil
}

func (p *SubtitleTrackHandler) HandleBinary(id mkvparse.ElementID, value []byte, info mkvparse.ElementInfo) error {
	switch id {
	case mkvparse.SeekIDElement:
		glog.V(10).Infof("%s- %v: %x\n", indent(info.Level), mkvparse.NameForElementID(id), value)
	default:
		glog.V(10).Infof("%s- %v: <binary> (%d)\n", indent(info.Level), mkvparse.NameForElementID(id), info.Size)
	}
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
	err := mkvparse.ParseSections(file, &h, mkvparse.TrackEntryElement, mkvParse.TracksElement)
	if err != nil {
		glog.Error(err)
	}
	return h.Subtitles, nil
}
