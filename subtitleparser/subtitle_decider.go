package subtitleparser

import (
	"errors"

	"github.com/golang/glog"
)

func DecideSubtitleTrack(subs map[int64]SubtitleTrack) (*SubtitleTrack, error) {
	//var track int64
	if len(subs) == 1 {
		glog.V(4).Info("Found one subtitle. Going to extract this one. Easy.")
		for _, value := range subs {
			if value.Codec == SSA || value.Codec == ASS {
				return &value, nil
			}
		}
	} else if len(subs) < 1 {
		return nil, errors.New("No subtitle tracks to extract")
	} else if hasDefault, defaultTrack := GetDefaultTrack(subs); hasDefault {
		glog.V(4).Infof("Found the default subtitle track %v, extracting this one.", defaultTrack.TrackID)
		return defaultTrack, nil
	} else if hasForced, forcedTrack := GetForcedTrack(subs); hasForced {
		glog.V(4).Infof("Found a forced subtitle track %v, extracting this one.", forcedTrack.TrackID)
	}
	return nil, errors.New("Could not find a suitable track to extract")
}

func GetDefaultTrack(subs map[int64]SubtitleTrack) (bool, *SubtitleTrack) {
	hasDefaultTrack := false
	var defaultTrack *SubtitleTrack
	defaultTrack = nil
	for _, value := range subs {
		if value.Default && !hasDefaultTrack {
			hasDefaultTrack = value.Default
			defaultTrack = &value
		} else if value.Default {
			glog.Warningf("Found 2 default tracks, %v and %v. Not sure how this happended, but I guess it does sometimes.", defaultTrack.TrackID, value.TrackID)
			glog.Warningf("I am taking this default track, since I saw it first: %v", defaultTrack.TrackID)
		}
	}
	return hasDefaultTrack, defaultTrack
}

func GetForcedTrack(subs map[int64]SubtitleTrack) (bool, *SubtitleTrack) {
	hasForcedTrack := false
	var forcedTrack *SubtitleTrack
	forcedTrack = nil
	for _, value := range subs {
		if value.Default && !hasForcedTrack {
			hasForcedTrack = value.Forced
			forcedTrack = &value
		} else if value.Forced {
			glog.Warningf("Found 2 forced tracks, %v and %v. Not sure how this happended, but I guess it does sometimes.", forcedTrack.TrackID, value.TrackID)
			glog.Warningf("I am taking this default track, since I saw it first: %v", forcedTrack.TrackID)
		}
	}
	return hasForcedTrack, forcedTrack
}
