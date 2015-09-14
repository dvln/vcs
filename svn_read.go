package vcs

import (
	"os/exec"
)

// SvnReader implements the Repo interface for the Svn source control.
type SvnReader struct {
	Description
}

// NewSvnReader creates a new instance of SvnReader. The remote and local directories
// need to be passed in. The remote location should include the branch for SVN.
// For example, if the package is https://github.com/Masterminds/cookoo/ the remote
// should be https://github.com/Masterminds/cookoo/trunk for the trunk branch.
func NewSvnReader(remote, wkspc string) (*SvnReader, error) {
	ltype, err := DetectVcsFromFS(wkspc)

	// Found a VCS other than Svn. Need to report an error.
	if err == nil && ltype != Svn {
		return nil, ErrWrongVCS
	}

	r := &SvnReader{}
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

// Update support for svn reader
func (r *SvnReader) Update(rev ...Rev) (string, error) {
	return SvnUpdate(r, rev...)
}

// Get support for svn reader
func (r *SvnReader) Get(rev ...Rev) (string, error) {
	return SvnGet(r, rev...)
}

// RevSet support for svn reader
func (r *SvnReader) RevSet(rev Rev) (string, error) {
	return SvnRevSet(r, rev)
}

// RevRead support for svn reader
func (r *SvnReader) RevRead(scope ...ReadScope) (*Revision, string, error) {
	return SvnRevRead(r, scope...)
}

// Exists support for svn reader
func (r *SvnReader) Exists(l Location) (bool, error) {
	return SvnExists(r, l)
}

