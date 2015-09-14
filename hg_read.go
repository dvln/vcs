package vcs

import (
	"os"
	"os/exec"
)

// HgReader implements the Repo interface for the Mercurial source control.
type HgReader struct {
	Description
}

// NewHgReader creates a new instance of HgReader. The remote and wkspc directories
// need to be passed in.
func NewHgReader(remote, wkspc string) (*HgReader, error) {
	ltype, err := DetectVcsFromFS(wkspc)

	// Found a VCS other than Hg. Need to report an error.
	if err == nil && ltype != Hg {
		return nil, ErrWrongVCS
	}

	r := &HgReader{}
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

// Update support for hg reader
func (r *HgReader) Update(rev ...Rev) (string, error) {
	return HgUpdate(r, rev...)
}

// Get support for hg reader
func (r *HgReader) Get(rev ...Rev) (string, error) {
	return HgGet(r, rev...)
}

// RevSet support for hg reader
func (r *HgReader) RevSet(rev Rev) (string, error) {
	return HgRevSet(r, rev)
}

// RevRead support for hg reader
func (r *HgReader) RevRead(scope ...ReadScope) (*Revision, string, error) {
	return HgRevRead(r, scope...)
}

// Exists support for hg reader
func (r *HgReader) Exists(l Location) (bool, error) {
	return HgExists(r, l)
}

