package vcs

// SvnUpdater implements the Repo interface for the Svn source control.
type SvnUpdater struct {
	Description
	Results
	mirror bool
	rebase RebaseVal
	refs   map[string]RefOp
}

// NewSvnUpdater creates a new instance of SvnUpdater. The remote and local directories
// need to be passed in. The remote location should include the branch for SVN.
// For example, if the package is https://github.com/Masterminds/cookoo/ the remote
// should be https://github.com/Masterminds/cookoo/trunk for the trunk branch.  Params:
//	remote (string): URL of remote repo
//	remoteName (string): If there is a name for the remote repo (SVN: use "", not used)
//	localPath (string): Directory for the local repo/clone/workspace to update
//	mirror (bool): if a full mirroring of all content is desired
//	rebase (RebaseVal): if rebase wanted or not, what type
//	refs (map[string]RefOp): list of refs to act on w/given operation (or nil)
// Currently ignores mirror, rebase and refs.
func NewSvnUpdater(remote, remoteName, localPath string, mirror bool, rebase RebaseVal, refs map[string]RefOp) (Updater, error) {
	ltype, err := DetectVcsFromFS(localPath)
	// Found a VCS other than Svn. Need to report an error.
	if err == nil && ltype != Svn {
		return nil, ErrWrongVCS
	}
	u := &SvnUpdater{}
	u.mirror = mirror
	u.rebase = rebase
	if refs != nil { // if refs given, then set up refs to act on w/ops
		u.refs = make(map[string]RefOp)
		u.refs = refs
	}
	u.setDescription(remote, remoteName, localPath, defaultSvnSchemes, Svn)
	if err == nil { // Have a localPath FS repo, try to validate/upd remote
		remote, _, err = SvnCheckRemote(u, remote)
		if err != nil {
			return nil, err
		}
		u.setRemote(remote)
	}
	return u, nil
}

// Update support for svn updater
func (u *SvnUpdater) Update(rev ...Rev) (Resulter, error) {
	return SvnUpdate(u, rev...)
}

// Exists support for svn updater
func (u *SvnUpdater) Exists(l Location) (string, Resulter, error) {
	return SvnExists(u, l)
}
