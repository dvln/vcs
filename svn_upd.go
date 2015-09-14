package vcs

import (
	"os/exec"
)

// SvnUpdater implements the Repo interface for the Svn source control.
type SvnUpdater struct {
	Description
}

// NewSvnUpdater creates a new instance of SvnUpdater. The remote and local directories
// need to be passed in. The remote location should include the branch for SVN.
// For example, if the package is https://github.com/Masterminds/cookoo/ the remote
// should be https://github.com/Masterminds/cookoo/trunk for the trunk branch.
func NewSvnUpdater(remote, wkspc string) (Updater, error) {
	ltype, err := DetectVcsFromFS(wkspc)

	// Found a VCS other than Svn. Need to report an error.
	if err == nil && ltype != Svn {
		return nil, ErrWrongVCS
	}

	r := &SvnUpdater{}
	r.setRemote(remote)
	r.setWkspcPath(wkspc)
	r.setVcs(Svn)

	// Make sure the wkspc SVN repo is configured the same as the remote when
	// A remote value was passed in.
	if exists, chkErr := r.Exists(Wkspc); err == nil && chkErr == nil && exists {
		// An SVN repo was found so test that the URL there matches
		// the repo passed in here.
		output, err := exec.Command("svn", "info", wkspc).CombinedOutput()
		if err != nil {
			return nil, err
		}

		m := svnDetectURL.FindStringSubmatch(string(output))
		if m[1] != "" && m[1] != remote {
			return nil, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Svn repo use that one.
		if remote == "" && m[1] != "" {
			r.setRemote(m[1])
		}
	}

	return r, nil
}

// Update support for svn updater
func (u *SvnUpdater) Update(rev ...Rev) (string, error) {
	return SvnUpdate(u, rev...)
}

// Exists support for svn updater
func (u *SvnUpdater) Exists(l Location) (bool, error) {
	return SvnExists(u, l)
}

