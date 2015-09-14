package vcs

import (
	"os"
	"os/exec"
)

// HgGetter implements the Repo interface for the Mercurial source control.
type HgGetter struct {
	Description
}

// NewHgGetter creates a new instance of HgGetter. The remote and wkspc directories
// need to be passed in.
func NewHgGetter(remote, wkspc string) (Getter, error) {
	ltype, err := DetectVcsFromFS(wkspc)

	// Found a VCS other than Hg. Need to report an error.
	if err == nil && ltype != Hg {
		return nil, ErrWrongVCS
	}

	r := &HgGetter{}
	r.setRemote(remote)
	r.setWkspcPath(wkspc)
	r.setVcs(Hg)

	// Make sure the wkspc Hg repo is configured the same as the remote when
	// A remote value was passed in.
	if exists, chkErr := r.Exists(Wkspc); err == nil && chkErr == nil && exists {
		// An Hg repo was found so test that the URL there matches
		// the repo passed in here.
		oldDir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		os.Chdir(wkspc)
		//FIXME: erik: this should be checked
		defer os.Chdir(oldDir)
		output, err := exec.Command("hg", "paths").CombinedOutput()
		if err != nil {
			return nil, err
		}

		m := hgDetectURL.FindStringSubmatch(string(output))
		if m[1] != "" && m[1] != remote {
			return nil, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Hg repo use that one.
		if remote == "" && m[1] != "" {
			r.setRemote(m[1])
		}
	}

	return r, nil
}

// Get support for hg getter
func (g *HgGetter) Get(rev ...Rev) (string, error) {
	return HgGet(g, rev...)
}

// RevSet support for hg getter
func (g *HgGetter) RevSet(rev Rev) (string, error) {
	return HgRevSet(g, rev)
}

// Exists support for hg getter
func (g *HgGetter) Exists(l Location) (bool, error) {
	return HgExists(g, l)
}

