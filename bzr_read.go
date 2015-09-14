package vcs

import (
	"os"
	"os/exec"
)

// BzrReader implements the Repo interface for the Bzr source control.
type BzrReader struct {
	Description
}

// NewBzrReader creates a new instance of BzrReader. The remote and wkspc directories
// need to be passed in.
func NewBzrReader(remote, wkspc string) (*BzrReader, error) {
	ltype, err := DetectVcsFromFS(wkspc)

	// Found a VCS other than Bzr. Need to report an error.
	if err == nil && ltype != Bzr {
		return nil, ErrWrongVCS
	}

	r := &BzrReader{}
	r.setRemote(remote)
	r.setWkspcPath(wkspc)
	r.setVcs(Bzr)

	// With the other VCS we can check if the endpoint locally is different
	// from the one configured internally. But, with Bzr you can't. For example,
	// if you do `bzr branch https://launchpad.net/govcstestbzrrepo` and then
	// use `bzr info` to get the parent branch you'll find it set to
	// http://bazaar.launchpad.net/~mattfarina/govcstestbzrrepo/trunk/. Notice
	// the change from https to http and the path chance.
	// Here we set the remote to be the local one if none is passed in.
	if exists, chkErr := r.Exists(Wkspc); err == nil && chkErr == nil && exists {
		oldDir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		os.Chdir(wkspc)
		if err != nil {
			return nil, err
		}
		defer os.Chdir(oldDir)
		output, err := exec.Command("bzr", "info").CombinedOutput()
		if err != nil {
			return nil, err
		}
		m := bzrDetectURL.FindStringSubmatch(string(output))

		// If no remote was passed in but one is configured for the locally
		// checked out Bzr repo use that one.
		if m[1] != "" {
			r.setRemote(m[1])
		}
	}

	return r, nil
}

// Update support for bzr reader
func (r *BzrReader) Update(rev ...Rev) (string, error) {
	return BzrUpdate(r, rev...)
}

// Get support for bzr reader
func (r *BzrReader) Get(rev ...Rev) (string, error) {
	return BzrGet(r, rev...)
}

// RevSet support for bzr reader
func (r *BzrReader) RevSet(rev Rev) (string, error) {
	return BzrRevSet(r, rev)
}

// RevRead support for bzr reader
func (r *BzrReader) RevRead(scope ...ReadScope) (*Revision, string, error) {
	return BzrRevRead(r, scope...)
}

// Exists support for bzr reader
func (r *BzrReader) Exists(l Location) (bool, error) {
	return BzrExists(r, l)
}

