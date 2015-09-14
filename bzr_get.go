package vcs

import (
	"os"
	"os/exec"
)

// BzrGetter implements the Repo interface for the Bzr source control.
type BzrGetter struct {
	Description
}

// NewBzrGetter creates a new instance of BzrGetter. The remote and wkspc
// directories need to be passed in.
func NewBzrGetter(remote, wkspc string) (Getter, error) {
	ltype, err := DetectVcsFromFS(wkspc)

	// Found a VCS other than Bzr. Need to report an error.
	if err == nil && ltype != Bzr {
		return nil, ErrWrongVCS
	}

	r := &BzrGetter{}
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
		err = os.Chdir(wkspc)
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

// Get support for bzr getter
func (g *BzrGetter) Get(rev ...Rev) (string, error) {
	return BzrGet(g, rev...)
}

// RevSet support for bzr getter
func (g *BzrGetter) RevSet(rev Rev) (string, error) {
	return BzrRevSet(g, rev)
}

// Exists support for bzr getter
func (g *BzrGetter) Exists(l Location) (bool, error) {
	return BzrExists(g, l)
}

