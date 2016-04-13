package vcs

// BzrUpdater implements the Repo interface for the Bzr source control.
type BzrUpdater struct {
	Description
	mirror bool
	rebase RebaseVal
	refs   map[string]RefOp
}

// NewBzrUpdater creates a new instance of BzrUpdater. The remote and wkspc
// directories need to be passed in.  Params:
//	remote (string): URL of remote repo
//	wkspc (string): Directory for the local workspace
//	mirror (bool): if a full mirroring of all content is desired
//	rebase (RebaseVal): if rebase wanted or not, what type
//	refs (map[string]RefOp): list of refs to act on w/given operation (or nil)
// Currently ignores mirror, rebase and refs.
func NewBzrUpdater(remote, wkspc string, mirror bool, rebase RebaseVal, refs map[string]RefOp) (Updater, error) {
	ltype, err := DetectVcsFromFS(wkspc)
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
	u.setDescription(remote, "", wkspc, defaultBzrSchemes, Bzr)
	if err == nil { // Have a local wkspc FS repo, try to improve the remote..
		remote, _, err = BzrCheckRemote(u, remote)
		if err != nil {
			return nil, err
		}
		u.setRemote(remote)
	}
	return u, nil
}

// Update support for bzr updater
func (u *BzrUpdater) Update(rev ...Rev) (string, error) {
	return BzrUpdate(u, rev...)
}

// Exists support for bzr updater
func (u *BzrUpdater) Exists(l Location) (string, error) {
	return BzrExists(u, l)
}
