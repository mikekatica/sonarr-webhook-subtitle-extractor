package subtitleparser

import "errors"

func DecideSubtitleTrack(subs map[int64]SubtitleTrack) (*int64, error) {
	//var track int64
	if len(subs) == 1 {
		for _, value := range subs {
			if value.Codec == SSA || value.Codec == ASS {
				return &value.TrackID, nil
			}
		}
	} else if len(subs) < 1 {
		return nil, errors.New("No subtitle tracks to extract")
	}
	return nil, errors.New("Could not find a suitable track to extract")
}
