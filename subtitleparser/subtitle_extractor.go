package subtitleparser

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/golang/glog"
)

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

var SubtitleExtensionMap = map[string]string{
	"S_TEXT/SSA": "ssa",
	"S_TEXT/ASS": "ass",
}

func ExtractSubtitleFromFile(pathIn string, subtrack *SubtitleTrack) error {
	mkvPath, mkvFilename := filepath.Split(pathIn)
	mkvArr := strings.Split(mkvFilename, ".")
	glog.V(5).Infof("Resolved Filename to %v", mkvArr)
	mkvFilenameNoExtension := strings.Join(mkvArr[:len(mkvArr)-1], ".")
	glog.V(5).Infof("Subtrack: %v", subtrack)
	subFilename := mkvFilenameNoExtension + ".default." + SubtitleExtensionMap[subtrack.Codec]
	subTempPath := "/tmp/" + subFilename
	subFullpath := mkvPath + subFilename
	subPathArgs := fmt.Sprintf("%d:%s", subtrack.TrackID-1, subTempPath)
	args := []string{shellescape.Quote(pathIn), "tracks", shellescape.Quote(subPathArgs)}
	glog.V(5).Infof("Using mkvextract args: %v", args)
	cmd := exec.Command("mkvextract", args...)
	out, err := cmd.CombinedOutput()
	outStr := string(out)
	if err != nil {
		glog.Errorf("Could not extract subs from %v, %v", mkvFilename, err)
		glog.Infoln("Output from mkvextract:")
		glog.Infof("%v\n", outStr)
		return err
	}
	glog.Infoln("Output from mkvextract:")
	glog.Infof("%v\n", outStr)
	err = Copy(subTempPath, subFullpath)
	if err != nil {
		glog.Errorf("Could not move sub file from %v to %v: %v", subTempPath, subFullpath, err)
		return err
	}
	return nil
}
