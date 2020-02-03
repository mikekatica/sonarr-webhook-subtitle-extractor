package subtitleparser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
)

var SubtitleExtensionMap = map[string]string{
	"S_TEXT/SSA": "ssa",
	"S_TEXT/ASS": "ass",
}

func ExtractSubtitleFromFile(pathIn string, subtrack SubtitleTrack) error {
	mkvPath, mkvFilename := filepath.Split(pathIn)
	mkvArr := strings.Split(mkvPath, ".")
	mkvFilenameNoExtension := strings.Join(mkvArr[:len(mkvArr)-1], ".")
	subFilename := mkvFilenameNoExtension + ".default." + SubtitleExtensionMap[subtrack.Codec]
	subTempPath := "/tmp/" + subFilename
	subFullpath := mkvPath + subFilename
	subPathArgs := fmt.Sprintf("%d:%s", subtrack.TrackID, subTempPath)
	args := []string{pathIn, "tracks", subPathArgs}
	cmd := exec.Command("mkvextract", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		glog.Errorf("Could not extract subs from %v, %v", mkvFilename, err)
		return err
	}
	outStr := string(out)
	glog.Infoln("Output from mkvextract:")
	glog.Infof("%v\n", outStr)
	err = os.Rename(subTempPath, subFullpath)
	if err != nil {
		glog.Errorf("Could not move sub file from %v to %v: %v", subTempPath, subFullpath, err)
		return err
	}
	return nil
}
