package vcs

// BzrUpdater implements the Repo interface for the Bzr source control.
type BzrUpdater struct {
	Description
	Results
	mirror bool
	rebase RebaseVal
	refs   map[string]RefOp
}

// NewBzrUpdater creates a new instance of BzrUpdater. The remote and localPath
// directories need to be passed in.  Params:
//	remote (string): URL of remote repo
//	remoteName (string): If there is a name for the remote repo (SVN: use "", not used)
//	localPath (string): Directory for the local repo/clone/workspace to update
//	mirror (bool): if a full mirroring of all content is desired
//	rebase (RebaseVal): if rebase wanted or not, what type
//	refs (map[string]RefOp): list of refs to act on w/given operation (or nil)
// Currently ignores mirror, rebase and refs.
func NewBzrUpdater(remote, remoteName, localPath string, mirror bool, rebase RebaseVal, refs map[string]RefOp) (Updater, error) {
	ltype, err := DetectVcsFromFS(localPath)
	// Found a VCS other than Bzr. Need to report an error.
	if err == nil && ltype != Bzr {
		return nil, ErrWrongVCS
	}
	u := &BzrUpdater{}
	u.mirror = mirror
	u.rebase = rebase
	if refs != nil { // if refs given, then set up refs to act on w/ops
		u.refs = make(map[string]RefOp)
		u.refs = refs
	}
	u.setDescription(remote, remoteName, localPath, defaultBzrSchemes, Bzr)
	if err == nil { // Have a localPath FS repo, try to improve the remote..
		remote, _, err = BzrCheckRemote(u, remote)
		if err != nil {
			return nil, err
		}
		u.setRemote(remote)
	}
	return u, nil
}

// Update support for bzr updater
func (u *BzrUpdater) Update(rev ...Rev) (Resulter, error) {
	return BzrUpdate(u, rev...)
}

// Exists support for bzr updater
func (u *BzrUpdater) Exists(l Location) (string, Resulter, error) {
	return BzrExists(u, l)
}
