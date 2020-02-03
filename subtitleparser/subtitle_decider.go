package subtitleparser

import (
	"errors"

	"github.com/golang/glog"
)

func DecideSubtitleTrack(subs map[int64]SubtitleTrack) (*SubtitleTrack, error) {
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
	}
	return nil, errors.New("Could not find a suitable track to extract")
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
