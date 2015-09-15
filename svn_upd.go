package vcs

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
	u := &SvnUpdater{}
	u.setRemote(remote)
	u.setWkspcPath(wkspc)
	u.setVcs(Svn)
	if err == nil { // Have a local wkspc FS repo, try to validate/upd remote
		remote, _, err = SvnCheckRemote(u, remote)
		if err != nil {
			return nil, err
		}
		u.setRemote(remote)
	}
	return u, nil
}

// Update support for svn updater
func (u *SvnUpdater) Update(rev ...Rev) (string, error) {
	return SvnUpdate(u, rev...)
}

// Exists support for svn updater
func (u *SvnUpdater) Exists(l Location) (bool, error) {
	return SvnExists(u, l)
}

