package subtitleparser

import (
	"errors"

	"github.com/golang/glog"
	"golang.org/x/text/language"
)

type SubtitleLanguageDefault struct {
	DefaultLang *string
}

func DecideSubtitleTrack(lang SubtitleLanguageDefault, subs map[int64]SubtitleTrack) (*SubtitleTrack, error) {
	//var track int64
	if len(subs) == 1 {
		glog.V(4).Info("Found one subtitle. Going to extract this one. Easy.")
		for key, value := range subs {
			if value.Codec == SSA || value.Codec == ASS {
				outSub := subs[key]
				return &outSub, nil
			}
			glog.V(4).Info("Subtitle track is not ASS or SSA, not extracting.")
		}
	} else if len(subs) < 1 {
		return nil, errors.New("No subtitle tracks to extract")
	} else if hasDefault, defaultTrack := GetDefaultTrack(subs); hasDefault && defaultTrack != nil {
		glog.V(4).Infof("Found the default subtitle track %v, extracting this one.", defaultTrack.TrackID)
		return defaultTrack, nil
	} else if hasForced, forcedTrack := GetForcedTrack(subs); hasForced && forcedTrack != nil {
		glog.V(4).Infof("Found a forced subtitle track %v, extracting this one.", forcedTrack.TrackID)
		return forcedTrack, nil
	} else if lang.DefaultLang != nil{
		if hasMatching, matchingTrack := GetLangMatchingTrack(*lang.DefaultLang, subs); hasMatching && matchingTrack != nil {
			glog.V(4).Infof("Found a language matching subtitle track %v, extracting this one.", matchingTrack.TrackID)
			return matchingTrack, nil
		}
	}
	return nil, errors.New("Could not find a suitable track to extract")
}

func GetLangMatchingTrack(lang string, subs map[int64]SubtitleTrack) (bool, *SubtitleTrack) {
	availableLangs := []language.Tag{}
	keys := []int64{}
	for key, value := range subs {
		ltag, err := language.Parse(value.Language)
		if err == nil {
			keys = append(keys, key)
			availableLangs = append(availableLangs, ltag)
		} else {
			glog.V(4).Infof("Could not process language %v, no match found", value.Language)
		}
	}
	matcher := language.NewMatcher(availableLangs)
	langTag, idx, conf := matcher.Match(language.MustParse(lang))
	glog.V(3).Infof("A Language match was found, '%v' with %v confidence", langTag, conf)
	if conf >= language.High {
		track := subs[keys[idx]]
		return true, &track
	} else {
		glog.V(4).Infof("Only High and Exact matches are considered")
		return false, nil
	}
}

func GetDefaultTrack(subs map[int64]SubtitleTrack) (bool, *SubtitleTrack) {
	hasDefaultTrack := false
	var defaultTrack *SubtitleTrack
	defaultTrack = nil
	for key, value := range subs {
		if value.Default && !hasDefaultTrack {
			hasDefaultTrack = value.Default
			outSub := subs[key]
			defaultTrack = &outSub
		} else if value.Default {
			glog.V(4).Infof("Found 2 default tracks, %v and %v. Not sure how this happended, but I guess it does sometimes.", defaultTrack.TrackID, value.TrackID)
			glog.V(4).Infof("I am assuming this default track, since I saw it first: %v", defaultTrack.TrackID)
		}
	}
	if defaultTrack != nil && defaultTrack.Codec != SSA && defaultTrack.Codec != ASS {
		glog.V(4).Info("Subtitle track is not ASS or SSA, not extracting.")
		return false, nil
	}
	return hasDefaultTrack, defaultTrack
}

func GetForcedTrack(subs map[int64]SubtitleTrack) (bool, *SubtitleTrack) {
	hasForcedTrack := false
	var forcedTrack *SubtitleTrack
	forcedTrack = nil
	for key, value := range subs {
		if value.Forced && !hasForcedTrack {
			hasForcedTrack = value.Forced
			outSub := subs[key]
			forcedTrack = &outSub
		} else if value.Forced {
			glog.V(4).Infof("Found 2 forced tracks, %v and %v. Not sure how this happended, but I guess it does sometimes.", forcedTrack.TrackID, value.TrackID)
			glog.V(4).Infof("I am taking this default track, since I saw it first: %v", forcedTrack.TrackID)
		}
	}
	if forcedTrack != nil && forcedTrack.Codec != SSA && forcedTrack.Codec != ASS {
		glog.V(4).Info("Subtitle track is not ASS or SSA, not extracting.")
		return false, nil
	}
	return hasForcedTrack, forcedTrack
}

func GetSubtitleTrack(subs map[int64]SubtitleTrack, track int64) (*SubtitleTrack, error) {
	//var track int64
	if len(subs) == 1 {
		glog.V(4).Info("Found one subtitle. Going to extract this one. Easy.")
		for key, value := range subs {
			if value.Codec == SSA || value.Codec == ASS {
				outSub := subs[key]
				return &outSub, nil
			}
			glog.V(4).Info("Subtitle track is not ASS or SSA, not extracting.")
		}
	} else if len(subs) < 1 {
		return nil, errors.New("No subtitle tracks to extract")
	} else {
		for key, value := range subs {
			if (value.Codec == SSA || value.Codec == ASS) && value.TrackID == track {
				outSub := subs[key]
				return &outSub, nil
			}
		}
	}
	return nil, errors.New("Could not find a suitable track to extract")
}
